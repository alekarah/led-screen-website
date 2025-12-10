package handlers

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

// ThumbnailSize определяет размеры thumbnails
type ThumbnailSize struct {
	Name   string
	Width  int
	Height int
	Suffix string
}

var (
	// ThumbnailSmall - для карточек проектов, главной, админки (~50KB)
	ThumbnailSmall = ThumbnailSize{
		Name:   "small",
		Width:  400,
		Height: 300,
		Suffix: "_small",
	}

	// ThumbnailMedium - для модального окна галереи (~180KB)
	ThumbnailMedium = ThumbnailSize{
		Name:   "medium",
		Width:  1200,
		Height: 900,
		Suffix: "_medium",
	}

	// AllThumbnailSizes - все размеры для генерации
	AllThumbnailSizes = []ThumbnailSize{
		ThumbnailSmall,
		ThumbnailMedium,
	}
)

// CropParams параметры кроппинга изображения
type CropParams struct {
	X     float64 // 0-100%
	Y     float64 // 0-100%
	Scale float64 // 0.5-3.0
}

// GenerateThumbnails создает все thumbnails для изображения с учетом кроппинга
// Возвращает map[suffix]path для сохранения в БД
func GenerateThumbnails(originalPath string, crop CropParams) (map[string]string, error) {
	// Открываем оригинальное изображение
	srcImage, err := imaging.Open(originalPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть изображение: %w", err)
	}

	// Применяем кроппинг если нужно
	if crop.Scale != 1.0 || crop.X != 50 || crop.Y != 50 {
		srcImage = applyCrop(srcImage, crop)
	}

	// Определяем anchor для Fill на основе crop параметров
	anchor := getAnchorForCrop(crop)

	// Генерируем thumbnails всех размеров
	thumbnailPaths := make(map[string]string)

	for _, size := range AllThumbnailSizes {
		thumbPath, err := generateSingleThumbnail(srcImage, originalPath, size, anchor)
		if err != nil {
			log.Printf("[WARN] Не удалось создать thumbnail %s: %v", size.Name, err)
			continue
		}
		thumbnailPaths[size.Suffix] = thumbPath
		log.Printf("[DEBUG] Создан thumbnail %s: %s", size.Name, thumbPath)
	}

	return thumbnailPaths, nil
}

// generateSingleThumbnail создает один thumbnail заданного размера
func generateSingleThumbnail(srcImage image.Image, originalPath string, size ThumbnailSize, anchor imaging.Anchor) (string, error) {
	// Создаем thumbnail с заполнением всего контейнера (эквивалент object-fit: cover)
	// Fill обрезает изображение так, чтобы оно заполнило весь размер без чёрных полос
	// Anchor определяет какую часть изображения сохранить при обрезке
	thumb := imaging.Fill(srcImage, size.Width, size.Height, anchor, imaging.Lanczos)

	// Генерируем имя файла для thumbnail (всегда .webp)
	thumbPath := generateThumbnailPath(originalPath, size.Suffix)

	// Сохраняем thumbnail как WebP
	if err := saveThumbnail(thumb, thumbPath); err != nil {
		return "", err
	}

	return thumbPath, nil
}

// applyCrop применяет параметры кроппинга к изображению
// Результат точно соответствует тому что показывает CSS preview в контейнере 4:3
func applyCrop(img image.Image, crop CropParams) image.Image {
	bounds := img.Bounds()
	imgWidth := float64(bounds.Dx())
	imgHeight := float64(bounds.Dy())

	// Целевое соотношение сторон контейнера (4:3)
	targetRatio := 4.0 / 3.0
	imgRatio := imgWidth / imgHeight

	// Вычисляем масштаб для object-fit: cover
	// Изображение должно полностью заполнить контейнер 4:3
	var coverScale float64
	if imgRatio > targetRatio {
		// Широкое изображение - высота заполняет контейнер
		coverScale = 1.0 / imgHeight // нормализуем к высоте 1
	} else {
		// Высокое изображение - ширина заполняет контейнер
		coverScale = targetRatio / imgWidth // нормализуем к ширине targetRatio
	}

	// Общий масштаб с учетом zoom пользователя
	totalScale := coverScale * crop.Scale

	// Размер видимой области в пикселях изображения
	// Контейнер 4:3, нормализованный как targetRatio x 1
	visibleWidth := targetRatio / totalScale
	visibleHeight := 1.0 / totalScale

	// CSS translate в процентах от размера изображения
	translateX := (crop.X - 50) * 2 / 100 * imgWidth
	translateY := (crop.Y - 50) * 2 / 100 * imgHeight

	// Центр видимой области (без translate - центр изображения)
	// translate сдвигает изображение, поэтому видимый центр сдвигается в противоположную сторону
	centerX := imgWidth/2 - translateX
	centerY := imgHeight/2 - translateY

	// Для вертикальных изображений корректируем translateY с учетом scale
	// CSS transform: scale() translate() - корректируем для совпадения с превью
	if imgRatio < targetRatio {
		// Вертикальное изображение - делим на scale чтобы сдвинуть вниз
		translateY = (crop.Y - 50) * 2 / 100 * imgHeight / crop.Scale
		centerY = imgHeight/2 - translateY
	}

	// Вычисляем координаты кропа
	x0 := int(centerX - visibleWidth/2)
	y0 := int(centerY - visibleHeight/2)
	x1 := int(centerX + visibleWidth/2)
	y1 := int(centerY + visibleHeight/2)

	log.Printf("[DEBUG] applyCrop: original=%dx%d, crop=(%.1f%%, %.1f%%, %.1fx), center=(%.0f,%.0f), visible=%.0fx%.0f, rect=(%d,%d)-(%d,%d)",
		int(imgWidth), int(imgHeight), crop.X, crop.Y, crop.Scale, centerX, centerY, visibleWidth, visibleHeight, x0, y0, x1, y1)

	// Ограничиваем координаты границами изображения
	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}
	if x1 > int(imgWidth) {
		x1 = int(imgWidth)
	}
	if y1 > int(imgHeight) {
		y1 = int(imgHeight)
	}

	// Применяем кроп
	cropRect := image.Rect(x0, y0, x1, y1)
	result := imaging.Crop(img, cropRect)
	log.Printf("[DEBUG] applyCrop: result=%dx%d", result.Bounds().Dx(), result.Bounds().Dy())
	return result
}

// getAnchorForCrop возвращает anchor для imaging.Fill на основе позиции кропа
func getAnchorForCrop(crop CropParams) imaging.Anchor {
	// Вычисляем видимый центр по формуле CSS
	visibleCenterY := 150.0 - 2.0*crop.Y

	// Выбираем anchor на основе вертикальной позиции
	// visibleCenterY < 40 означает что пользователь смотрит на верхнюю часть изображения
	// visibleCenterY > 60 означает что пользователь смотрит на нижнюю часть
	if visibleCenterY < 40 {
		return imaging.Top // Сохраняем верх
	} else if visibleCenterY > 60 {
		return imaging.Bottom // Сохраняем низ
	}
	return imaging.Center
}

// saveThumbnail сохраняет изображение с оптимизацией
// Все thumbnails сохраняются в формате WebP с качеством 90%
func saveThumbnail(img image.Image, path string) error {
	// Создаем директорию если нужно
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию: %w", err)
	}

	// Создаем файл
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("не удалось создать файл: %w", err)
	}
	defer file.Close()

	// Сохраняем как WebP с качеством 90% (lossy)
	// Quality 90 обеспечивает визуально неотличимое качество при экономии 25-35% размера
	options := &webp.Options{
		Lossless: false,
		Quality:  90,
	}

	if err := webp.Encode(file, img, options); err != nil {
		return fmt.Errorf("не удалось сохранить WebP: %w", err)
	}

	return nil
}

// generateThumbnailPath генерирует путь для thumbnail
// Всегда использует .webp расширение для thumbnails
func generateThumbnailPath(originalPath, suffix string) string {
	ext := filepath.Ext(originalPath)
	nameWithoutExt := strings.TrimSuffix(originalPath, ext)
	return nameWithoutExt + suffix + ".webp"
}

// DeleteThumbnails удаляет все thumbnails для изображения
func DeleteThumbnails(originalPath string) {
	for _, size := range AllThumbnailSizes {
		// Удаляем WebP thumbnails (новый формат)
		thumbPath := generateThumbnailPath(originalPath, size.Suffix)
		if err := os.Remove(thumbPath); err != nil && !os.IsNotExist(err) {
			log.Printf("[WARN] Не удалось удалить thumbnail %s: %v", thumbPath, err)
		}

		// Также удаляем старые PNG/JPG thumbnails если они есть
		ext := filepath.Ext(originalPath)
		nameWithoutExt := strings.TrimSuffix(originalPath, ext)
		oldPngPath := nameWithoutExt + size.Suffix + ext
		if err := os.Remove(oldPngPath); err != nil && !os.IsNotExist(err) {
			log.Printf("[WARN] Не удалось удалить старый thumbnail %s: %v", oldPngPath, err)
		}
	}
}
