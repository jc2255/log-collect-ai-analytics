package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cj/log-collect-ai-analytics/internal/pkg/redis"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
)

const captchaRedisPrefix = "lca:captcha:"
const captchaTTL = 5 * time.Minute

// CaptchaHandler 验证码
// 优先使用 Redis 存储（HA 多实例共享）；Redis 不可用时回退到内存。
type CaptchaHandler struct {
	store sync.Map // key -> captchaEntry，回退用
}

type captchaEntry struct {
	Code      string
	ExpiredAt time.Time
}

func NewCaptchaHandler() *CaptchaHandler {
	return &CaptchaHandler{}
}

// Generate 生成验证码
func (h *CaptchaHandler) Generate(c *gin.Context) {
	// 生成4位数字验证码
	code := fmt.Sprintf("%04d", rand.Intn(10000))
	key := fmt.Sprintf("%d%d", time.Now().UnixNano(), rand.Intn(1000))

	// 优先存 Redis，失败时回退本地内存
	if rdb := redis.GetClient(); rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := rdb.Set(ctx, captchaRedisPrefix+key, code, captchaTTL).Err(); err != nil {
			h.store.Store(key, captchaEntry{Code: code, ExpiredAt: time.Now().Add(captchaTTL)})
		}
	} else {
		h.store.Store(key, captchaEntry{Code: code, ExpiredAt: time.Now().Add(captchaTTL)})
	}

	// 生成图片
	img := h.generateImage(code)
	var buf bytes.Buffer
	png.Encode(&buf, img)
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	response.Success(c, gin.H{
		"captcha_id":    key,
		"captcha_image": "data:image/png;base64," + b64,
	})
}

// Verify 验证验证码
func (h *CaptchaHandler) Verify(captchaID, captchaCode string) bool {
	// 先从 Redis 取（HA 多实例共享）
	if rdb := redis.GetClient(); rdb != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		rkey := captchaRedisPrefix + captchaID
		val, err := rdb.Get(ctx, rkey).Result()
		if err == nil {
			// 取出后立即删除，防止重放
			rdb.Del(ctx, rkey)
			return val == captchaCode
		}
		// Redis 中找不到，再尝试回退内存（本实例曾在 Redis 失败时写过）
	}

	val, ok := h.store.LoadAndDelete(captchaID)
	if !ok {
		return false
	}
	entry := val.(captchaEntry)
	if time.Now().After(entry.ExpiredAt) {
		return false
	}
	return entry.Code == captchaCode
}

// generateImage 生成简单的验证码图片
func (h *CaptchaHandler) generateImage(code string) image.Image {
	width, height := 120, 40
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 背景色
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{240, 240, 240, 255})
		}
	}

	// 添加干扰线
	for i := 0; i < 4; i++ {
		x1 := rand.Intn(width)
		y1 := rand.Intn(height)
		x2 := rand.Intn(width)
		y2 := rand.Intn(height)
		lineColor := color.RGBA{uint8(rand.Intn(200)), uint8(rand.Intn(200)), uint8(rand.Intn(200)), 255}
		drawLine(img, x1, y1, x2, y2, lineColor)
	}

	// 绘制数字 - 使用简单的像素字体
	colors := []color.RGBA{
		{50, 50, 150, 255},
		{150, 50, 50, 255},
		{50, 150, 50, 255},
		{150, 50, 150, 255},
	}
	for i, ch := range code {
		drawDigit(img, int(ch-'0'), 15+i*25, 8, colors[i%len(colors)])
	}

	// 添加噪点
	for i := 0; i < 100; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		img.Set(x, y, color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255})
	}

	return img
}

// drawLine Bresenham 画线
func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 >= x2 {
		sx = -1
	}
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy
	for {
		img.Set(x1, y1, c)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// 3x5 像素字体定义
var digitPatterns = [10][5][3]bool{
	// 0
	{{true, true, true}, {true, false, true}, {true, false, true}, {true, false, true}, {true, true, true}},
	// 1
	{{false, true, false}, {true, true, false}, {false, true, false}, {false, true, false}, {true, true, true}},
	// 2
	{{true, true, true}, {false, false, true}, {true, true, true}, {true, false, false}, {true, true, true}},
	// 3
	{{true, true, true}, {false, false, true}, {true, true, true}, {false, false, true}, {true, true, true}},
	// 4
	{{true, false, true}, {true, false, true}, {true, true, true}, {false, false, true}, {false, false, true}},
	// 5
	{{true, true, true}, {true, false, false}, {true, true, true}, {false, false, true}, {true, true, true}},
	// 6
	{{true, true, true}, {true, false, false}, {true, true, true}, {true, false, true}, {true, true, true}},
	// 7
	{{true, true, true}, {false, false, true}, {false, false, true}, {false, false, true}, {false, false, true}},
	// 8
	{{true, true, true}, {true, false, true}, {true, true, true}, {true, false, true}, {true, true, true}},
	// 9
	{{true, true, true}, {true, false, true}, {true, true, true}, {false, false, true}, {true, true, true}},
}

func drawDigit(img *image.RGBA, digit, offsetX, offsetY int, c color.RGBA) {
	scale := 5
	pattern := digitPatterns[digit]
	for row := 0; row < 5; row++ {
		for col := 0; col < 3; col++ {
			if pattern[row][col] {
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						img.Set(offsetX+col*scale+dx, offsetY+row*scale+dy, c)
					}
				}
			}
		}
	}
}
