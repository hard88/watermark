// Package watermark 提供一个简单的水印功能。
package watermark

import (
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ErrUnsupportedWatermarkType 不支持的水印类型
var ErrUnsupportedWatermarkType = errors.New("不支持的水印类型")

// 允许做水印的图片类型
var allowExts = []string{
	".jpg", ".jpeg", ".png",
}


// Watermark 用于给图片添加水印功能。
// 目前支持  png 三种图片格式。
// 若是 gif 图片，则只取图片的第一帧；png 支持透明背景。
type Watermark struct {
	image   image.Image // 水印图片
}

// New 声明一个 Watermark 对象。
//
// path 为水印文件的路径；
// padding 为水印在目标图像上的留白大小；
// pos 水印的位置。
func New(path string) (*Watermark, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var img image.Image
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(f)
	case ".png":
		img, err = png.Decode(f)
	default:
		return nil, ErrUnsupportedWatermarkType
	}
	if err != nil {
		return nil, err
	}

	return &Watermark{
		image:   img,
	}, nil
}

// IsAllowExt 该扩展名的图片是否允许使用水印
//
// ext 必须带上 . 符号
func IsAllowExt(ext string) bool {
	if ext == "" {
		panic("参数 ext 不能为空")
	}

	if ext[0] != '.' {
		panic("参数 ext 必须以 . 开头")
	}

	ext = strings.ToLower(ext)

	for _, e := range allowExts {
		if e == ext {
			return true
		}
	}
	return false
}

// MarkFile 给指定的文件打上水印
func (w *Watermark) MarkFile(path string, point image.Point) error {
	file, err := os.OpenFile(path, os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	return w.Mark(file, strings.ToLower(filepath.Ext(path)), point)
}

// Mark 将水印写入 src 中，由 ext 确定当前图片的类型。
func (w *Watermark) Mark(src io.ReadWriteSeeker, ext string, point image.Point) (err error) {
	var srcImg image.Image

	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg":
		srcImg, err = jpeg.Decode(src)
	case ".png":
		srcImg, err = png.Decode(src)
	default:
		return ErrUnsupportedWatermarkType
	}
	if err != nil {
		return err
	}

	dstImg := image.NewNRGBA64(srcImg.Bounds())
	draw.Draw(dstImg, dstImg.Bounds(), srcImg, image.ZP, draw.Src)
	draw.Draw(dstImg, dstImg.Bounds(), w.image, point, draw.Over)

	if _, err = src.Seek(0, 0); err != nil {
		return err
	}

	switch ext {
	case ".jpg", ".jpeg":
		return jpeg.Encode(src, dstImg, nil)
	case ".png":
		return png.Encode(src, dstImg)
	default:
		return ErrUnsupportedWatermarkType
	}
}
