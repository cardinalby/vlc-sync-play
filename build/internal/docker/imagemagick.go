package docker

import (
	"path/filepath"
	"strconv"
	"strings"
)

func imageMagicConvert(volumes map[string]string, args []string) error {
	return RunImage("dpokidov/imagemagick", volumes, args)
}

func imageMagickVolumes(srcPath, dstPath string) (
	vols map[string]string,
	srcInContainer,
	dstInContainer string,
) {
	return map[string]string{
			srcPath:               "/src_f",
			filepath.Dir(dstPath): "/dst",
		},
		"/src_f",
		"/dst/" + filepath.Base(dstPath)
}

func ImageMagicPsdToPng(srcPath, dstPath string) error {
	vols, src, dst := imageMagickVolumes(srcPath, dstPath)
	return imageMagicConvert(
		vols,
		[]string{
			src + "[0]",
			"PNG:" + dst,
		})
}

func ImageMagicPsdToIco(srcPath, dstPath string, icoSizes []int) error {
	var sizesStrArr []string
	for _, size := range icoSizes {
		sizesStrArr = append(sizesStrArr, strconv.Itoa(size))
	}
	sizesStr := strings.Join(sizesStrArr, ",")
	vols, src, dst := imageMagickVolumes(srcPath, dstPath)

	return imageMagicConvert(
		vols,
		[]string{
			src + "[0]",
			"-define",
			"icon:auto-resize=" + sizesStr,
			"ICO:" + dst,
		})
}

func ImageMagicPngToIco(srcPath, dstPath string, icoSizes []int) error {
	var sizesStrArr []string
	for _, size := range icoSizes {
		sizesStrArr = append(sizesStrArr, strconv.Itoa(size))
	}
	sizesStr := strings.Join(sizesStrArr, ",")
	vols, src, dst := imageMagickVolumes(srcPath, dstPath)

	return imageMagicConvert(
		vols,
		[]string{
			src,
			"-define",
			"icon:auto-resize=" + sizesStr,
			"ICO:" + dst,
		})
}
