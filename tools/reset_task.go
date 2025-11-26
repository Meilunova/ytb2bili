package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/difyz9/ytb2bili/internal/core/types"
	"github.com/difyz9/ytb2bili/pkg/store"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法:")
		fmt.Println("  go run reset_task.go <video_id>        # 重置指定视频的所有任务")
		fmt.Println("  go run reset_task.go <video_id> clean  # 重置任务并清理文件")
		fmt.Println("  go run reset_task.go all               # 重置所有失败的任务")
		fmt.Println("  go run reset_task.go all clean         # 重置所有失败任务并清理文件")
		os.Exit(1)
	}

	videoID := os.Args[1]
	cleanFiles := len(os.Args) > 2 && os.Args[2] == "clean"

	// 加载配置
	config, err := types.LoadConfig("config.toml")
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	// 连接数据库
	db, err := store.NewDatabase(config)
	if err != nil {
		log.Fatalf("❌ 连接数据库失败: %v", err)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🔄 任务重置工具")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if videoID == "all" {
		resetAllFailedTasks(db, cleanFiles)
	} else {
		resetVideoTasks(db, videoID, cleanFiles)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ 操作完成")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// resetVideoTasks 重置指定视频的任务
func resetVideoTasks(db *gorm.DB, videoID string, cleanFiles bool) {
	fmt.Printf("📹 视频ID: %s\n", videoID)

	// 查询视频信息
	var video struct {
		VideoID string
		Title   string
		Status  string
	}
	if err := db.Table("cw_saved_videos").
		Where("video_id = ?", videoID).
		First(&video).Error; err != nil {
		log.Fatalf("❌ 未找到视频: %v", err)
	}

	fmt.Printf("📝 标题: %s\n", video.Title)
	fmt.Printf("📊 当前状态: %s\n\n", video.Status)

	// 查询任务步骤
	var steps []struct {
		ID       uint
		StepName string
		Status   string
		ErrorMsg string
	}
	db.Table("cw_task_steps").
		Where("video_id = ?", videoID).
		Order("step_order").
		Find(&steps)

	fmt.Println("📋 任务步骤:")
	for _, step := range steps {
		status := "✅"
		if step.Status == "failed" {
			status = "❌"
		} else if step.Status == "pending" {
			status = "⏳"
		} else if step.Status == "running" {
			status = "🔄"
		}
		fmt.Printf("  %s %s - %s\n", status, step.StepName, step.Status)
		if step.ErrorMsg != "" {
			fmt.Printf("     错误: %s\n", step.ErrorMsg)
		}
	}

	fmt.Println("\n🔄 开始重置...")

	// 重置失败的任务步骤
	result := db.Table("cw_task_steps").
		Where("video_id = ? AND status = ?", videoID, "failed").
		Updates(map[string]interface{}{
			"status":     "pending",
			"error_msg":  nil,
			"start_time": nil,
			"end_time":   nil,
			"duration":   nil,
		})

	if result.Error != nil {
		log.Fatalf("❌ 重置任务失败: %v", result.Error)
	}

	fmt.Printf("✓ 已重置 %d 个失败的任务步骤\n", result.RowsAffected)

	// 重置视频状态
	if video.Status != "001" {
		db.Table("cw_saved_videos").
			Where("video_id = ?", videoID).
			Update("status", "001")
		fmt.Println("✓ 已重置视频状态为待处理 (001)")
	}

	// 清理文件
	if cleanFiles {
		cleanVideoFiles(videoID)
	}
}

// resetAllFailedTasks 重置所有失败的任务
func resetAllFailedTasks(db *gorm.DB, cleanFiles bool) {
	fmt.Println("🔍 查找所有失败的任务...")

	// 查询所有有失败任务的视频
	var videoIDs []string
	db.Table("cw_task_steps").
		Where("status = ?", "failed").
		Distinct("video_id").
		Pluck("video_id", &videoIDs)

	if len(videoIDs) == 0 {
		fmt.Println("✓ 没有失败的任务")
		return
	}

	fmt.Printf("📊 找到 %d 个视频有失败的任务\n\n", len(videoIDs))

	for i, videoID := range videoIDs {
		fmt.Printf("[%d/%d] ", i+1, len(videoIDs))
		resetVideoTasks(db, videoID, cleanFiles)
		fmt.Println()
	}
}

// cleanVideoFiles 清理视频相关文件
func cleanVideoFiles(videoID string) {
	fmt.Println("\n🧹 清理文件...")

	// 查找视频目录
	mediaDir := "data/media"
	var videoDir string

	// 遍历日期目录
	filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == videoID {
			videoDir = path
			return filepath.SkipDir
		}
		return nil
	})

	if videoDir == "" {
		fmt.Println("  ℹ️  未找到视频文件目录")
		return
	}

	fmt.Printf("  📁 找到目录: %s\n", videoDir)

	// 列出要删除的文件
	var filesToDelete []string
	filepath.Walk(videoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			// 保留字幕文件，删除其他文件
			ext := filepath.Ext(path)
			if ext != ".srt" && ext != ".vtt" {
				filesToDelete = append(filesToDelete, path)
			}
		}
		return nil
	})

	if len(filesToDelete) == 0 {
		fmt.Println("  ℹ️  没有需要清理的文件")
		return
	}

	fmt.Println("  📄 将删除以下文件:")
	for _, file := range filesToDelete {
		fmt.Printf("    - %s\n", filepath.Base(file))
	}

	// 删除文件
	deletedCount := 0
	for _, file := range filesToDelete {
		if err := os.Remove(file); err != nil {
			fmt.Printf("    ⚠️  删除失败: %s - %v\n", filepath.Base(file), err)
		} else {
			deletedCount++
		}
	}

	fmt.Printf("  ✓ 已删除 %d 个文件\n", deletedCount)

	// 如果目录为空（只剩字幕），可以选择保留或删除
	remaining, _ := os.ReadDir(videoDir)
	if len(remaining) == 0 {
		os.Remove(videoDir)
		fmt.Println("  ✓ 已删除空目录")
	}
}
