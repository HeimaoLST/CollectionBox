package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github/heimaolst/collectionbox/internal/biz"
	"github/heimaolst/collectionbox/internal/logx"
)

type CollectionService struct {
	uc *biz.CollectionUsecase
}

type CreateRequest struct {
	URL string `json:"url"`
}
type GetByTimeRangeRequest struct {
	Origin string     `json:"origin"`
	Start  *time.Time `json:"start,omitempty"`
	End    *time.Time `json:"end,omitempty"`
}

func NewService(uc *biz.CollectionUsecase) *CollectionService {
	return &CollectionService{
		uc: uc,
	}
}

func (s *CollectionService) CreateCollection(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	// 1. 解析请求
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON format: "+err.Error())
		return
	}
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	// 2. 调用 Biz 层 (现在的逻辑是：有则更新，无则创建)
	// 方法名建议改为 UpsertCollectionsFromText 或保持原样但修改内部逻辑
	ctx := r.Context()
	cols, err := s.uc.UpsertCollectionsFromText(ctx, req.URL)

	// 3. 错误处理
	if err != nil {
		// 如果有参数错误（如解析不出 URL）
		if errors.Is(err, biz.ErrInvalidArgument) {
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			logx.FromContext(ctx).Error("upsert collections failed", "err", err)
			writeError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// 4. 返回结果
	// Upsert 语义下，通常返回 200 OK，因为它不全是新建
	writeJSON(w, http.StatusOK, cols)
}

type UpdateCollectionTimeRequest struct {
	dups []string
}

func (s *CollectionService) UpdateCollectionTime(w http.ResponseWriter, r *http.Request) {
	var req UpdateCollectionTimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusInternalServerError, "json decode error")
	}
	err := s.uc.UpdateCollectionCreateTime(r.Context(), req.dups)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "UpdateCollectionTime error")
	}

	writeJSON(w, http.StatusOK, nil)
}

func (s *CollectionService) GetByOrigin(w http.ResponseWriter, r *http.Request) {
	targetOrigin := r.FormValue("origin")
	var (
		res interface{}
		err error
	)
	if targetOrigin != "" {
		cols, err := s.uc.GetByOrigin(r.Context(), targetOrigin)
		if err != nil {
			logx.FromContext(r.Context()).Error("get by origin failed", "err", err)
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		mp := make(map[string][]*biz.Collection)
		for _, col := range cols {
			mp[col.Origin] = append(mp[col.Origin], col)
		}
		res = mp
	} else {
		res, err = s.uc.GetAllGroupedByOrigin(r.Context())
	}
	if err != nil {
		logx.FromContext(r.Context()).Error("get by time range failed", "err", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *CollectionService) GetAll(w http.ResponseWriter, r *http.Request) {
}

func (s *CollectionService) GetByTimeRange(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}
	var req GetByTimeRangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON format: "+err.Error())
		return
	}
	// clean data
	// Set default time range to the last 24 hours if not provided
	// 没有参数视为一天
	if req.Start == nil {
		t := time.Now().Add(-24 * time.Hour)
		req.Start = &t
	}
	if req.End == nil {
		t := time.Now()
		req.End = &t
		// if req.Origin == "" it will return all origin (handled by biz layer)
		cols, err := s.uc.GetByTimeRange(r.Context(), *req.Start, *req.End, req.Origin)
		if err != nil {
			logx.FromContext(r.Context()).Error("get by time range failed", "err", err)
			writeError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		writeJSON(w, http.StatusOK, cols)

	}
}

// --- 辅助函数 (可以放在这个文件的末尾，或单独的包里) ---

func writeError(w http.ResponseWriter, statusCode int, message string) {
	type ErrorResponse struct {
		Error string `json:"error"`
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-F8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// no request context here; use base logger with TODO context
		logx.FromContext(context.TODO()).Error("encode response failed", "err", err)
	}
}
