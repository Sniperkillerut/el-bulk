package service

import (
	"context"

	"github.com/el-bulk/backend/store"
)

type SitemapData struct {
	Products    []string `json:"products"`
	Notices     []string `json:"notices"`
	TCGs        []string `json:"tcgs"`
	Collections []string `json:"collections"`
}

type SeoService struct {
	ProductStore  *store.ProductStore
	NoticeStore   *store.NoticeStore
	TCGStore      *store.TCGStore
	CategoryStore *store.CategoryStore
}

func NewSeoService(ps *store.ProductStore, ns *store.NoticeStore, ts *store.TCGStore, cs *store.CategoryStore) *SeoService {
	return &SeoService{
		ProductStore:  ps,
		NoticeStore:   ns,
		TCGStore:      ts,
		CategoryStore: cs,
	}
}

func (s *SeoService) GetSitemapData(ctx context.Context) (*SitemapData, error) {
	data := &SitemapData{
		Products:    []string{},
		Notices:     []string{},
		TCGs:        []string{},
		Collections: []string{},
	}

	// 1. Fetch Active Products (only IDs for indexing)
	// We use the store directly to bypass heavy joins if possible, 
	// but here we can just do a simple query.
	queryProducts := "SELECT id FROM product WHERE stock > 0"
	if err := s.ProductStore.DB.SelectContext(ctx, &data.Products, queryProducts); err != nil {
		return nil, err
	}

	// 2. Fetch Published Notice Slugs
	queryNotices := "SELECT slug FROM notice WHERE is_published = true"
	if err := s.NoticeStore.DB.SelectContext(ctx, &data.Notices, queryNotices); err != nil {
		return nil, err
	}

	// 3. Fetch Active TCG IDs
	queryTCGs := "SELECT id FROM tcg WHERE is_active = true"
	if err := s.TCGStore.DB.SelectContext(ctx, &data.TCGs, queryTCGs); err != nil {
		return nil, err
	}

	// 4. Fetch Active Category Slugs
	queryCollections := "SELECT slug FROM custom_category WHERE is_active = true"
	if err := s.CategoryStore.DB.SelectContext(ctx, &data.Collections, queryCollections); err != nil {
		return nil, err
	}

	return data, nil
}
