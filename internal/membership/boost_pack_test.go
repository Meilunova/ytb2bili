package membership

import (
	"context"
	"testing"
	"time"
)

func TestBoostPackService_GetStatus(t *testing.T) {
	store := NewMockStore()
	service := NewBoostPackService(store)
	ctx := context.Background()

	// 没有加油包
	status, err := service.GetStatus(ctx, "user1")
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}
	if status.HasPack {
		t.Error("User should not have pack initially")
	}
	if status.VideosRemaining != 0 {
		t.Errorf("Videos remaining should be 0, got %d", status.VideosRemaining)
	}

	// 添加加油包
	store.SetBoostPack("user1", 10, 7)
	status, _ = service.GetStatus(ctx, "user1")
	if !status.HasPack {
		t.Error("User should have pack after adding")
	}
	if status.VideosRemaining != 10 {
		t.Errorf("Videos remaining should be 10, got %d", status.VideosRemaining)
	}
	if status.DaysRemaining < 6 || status.DaysRemaining > 7 {
		t.Errorf("Days remaining should be ~7, got %d", status.DaysRemaining)
	}
}

func TestBoostPackService_GetStatus_Expired(t *testing.T) {
	store := NewMockStore()
	service := NewBoostPackService(store)
	ctx := context.Background()

	// 添加过期的加油包
	store.boostPacks["user1"] = &UserBoostPack{
		UserID:          "user1",
		VideosRemaining: 10,
		ExpiresAt:       time.Now().Add(-24 * time.Hour), // 已过期
	}

	status, _ := service.GetStatus(ctx, "user1")
	if status.HasPack {
		t.Error("Expired pack should not be valid")
	}
}

func TestBoostPackService_Purchase_New(t *testing.T) {
	store := NewMockStore()
	service := NewBoostPackService(store)
	ctx := context.Background()

	result, err := service.Purchase(ctx, "user1", BoostPackSmall)
	if err != nil {
		t.Fatalf("Purchase failed: %v", err)
	}

	if !result.Success {
		t.Error("Purchase should succeed")
	}
	if !result.NewPurchase {
		t.Error("Should be a new purchase")
	}
	if result.VideosAdded != 10 {
		t.Errorf("Videos added should be 10, got %d", result.VideosAdded)
	}
	if result.TotalVideos != 10 {
		t.Errorf("Total videos should be 10, got %d", result.TotalVideos)
	}

	// 验证存储
	pack, _ := store.GetUserBoostPack(ctx, "user1")
	if pack.VideosRemaining != 10 {
		t.Errorf("Stored pack should have 10 videos, got %d", pack.VideosRemaining)
	}
}

func TestBoostPackService_Purchase_Stack(t *testing.T) {
	store := NewMockStore()
	service := NewBoostPackService(store)
	ctx := context.Background()

	// 先购买一个小包
	service.Purchase(ctx, "user1", BoostPackSmall)

	// 再购买一个中包，应该叠加
	result, err := service.Purchase(ctx, "user1", BoostPackMedium)
	if err != nil {
		t.Fatalf("Second purchase failed: %v", err)
	}

	if result.NewPurchase {
		t.Error("Should not be a new purchase (stacking)")
	}
	if result.VideosAdded != 30 {
		t.Errorf("Videos added should be 30, got %d", result.VideosAdded)
	}
	if result.TotalVideos != 40 { // 10 + 30
		t.Errorf("Total videos should be 40, got %d", result.TotalVideos)
	}
}

func TestBoostPackService_Purchase_InvalidType(t *testing.T) {
	store := NewMockStore()
	service := NewBoostPackService(store)
	ctx := context.Background()

	_, err := service.Purchase(ctx, "user1", "invalid_type")
	if err == nil {
		t.Error("Purchase with invalid type should fail")
	}
}

func TestBoostPackService_Purchase_AllTypes(t *testing.T) {
	store := NewMockStore()
	service := NewBoostPackService(store)
	ctx := context.Background()

	tests := []struct {
		packType BoostPackType
		videos   int
	}{
		{BoostPackSmall, 10},
		{BoostPackMedium, 30},
		{BoostPackLarge, 80},
	}

	for _, tt := range tests {
		t.Run(string(tt.packType), func(t *testing.T) {
			store.Reset()
			userID := "user-" + string(tt.packType)

			result, err := service.Purchase(ctx, userID, tt.packType)
			if err != nil {
				t.Fatalf("Purchase %s failed: %v", tt.packType, err)
			}
			if result.VideosAdded != tt.videos {
				t.Errorf("Expected %d videos, got %d", tt.videos, result.VideosAdded)
			}
		})
	}
}

func TestBoostPackService_Consume(t *testing.T) {
	store := NewMockStore()
	service := NewBoostPackService(store)
	ctx := context.Background()

	store.SetBoostPack("user1", 3, 7)

	// 消耗
	err := service.Consume(ctx, "user1")
	if err != nil {
		t.Errorf("Consume failed: %v", err)
	}

	// 验证减少
	pack, _ := store.GetUserBoostPack(ctx, "user1")
	if pack.VideosRemaining != 2 {
		t.Errorf("Should have 2 remaining, got %d", pack.VideosRemaining)
	}

	// 消耗完
	service.Consume(ctx, "user1")
	service.Consume(ctx, "user1")

	// 再消耗应该失败
	err = service.Consume(ctx, "user1")
	if err == nil {
		t.Error("Consume should fail when pack is empty")
	}
}

func TestBoostPackService_HasValidPack(t *testing.T) {
	store := NewMockStore()
	service := NewBoostPackService(store)
	ctx := context.Background()

	// 没有加油包
	if service.HasValidPack(ctx, "user1") {
		t.Error("Should not have valid pack initially")
	}

	// 添加加油包
	store.SetBoostPack("user1", 10, 7)
	if !service.HasValidPack(ctx, "user1") {
		t.Error("Should have valid pack after adding")
	}

	// 过期的加油包
	store.boostPacks["user2"] = &UserBoostPack{
		UserID:          "user2",
		VideosRemaining: 10,
		ExpiresAt:       time.Now().Add(-time.Hour),
	}
	if service.HasValidPack(ctx, "user2") {
		t.Error("Expired pack should not be valid")
	}

	// 空的加油包
	store.boostPacks["user3"] = &UserBoostPack{
		UserID:          "user3",
		VideosRemaining: 0,
		ExpiresAt:       time.Now().Add(time.Hour),
	}
	if service.HasValidPack(ctx, "user3") {
		t.Error("Empty pack should not be valid")
	}
}

func TestBoostPackService_GetAvailableVideos(t *testing.T) {
	store := NewMockStore()
	service := NewBoostPackService(store)
	ctx := context.Background()

	// 没有加油包
	videos := service.GetAvailableVideos(ctx, "user1")
	if videos != 0 {
		t.Errorf("Should have 0 videos, got %d", videos)
	}

	// 有加油包
	store.SetBoostPack("user1", 15, 7)
	videos = service.GetAvailableVideos(ctx, "user1")
	if videos != 15 {
		t.Errorf("Should have 15 videos, got %d", videos)
	}
}
