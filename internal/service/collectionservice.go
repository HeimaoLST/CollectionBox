package service

import (
	"context"
	"encoding/json"
	"errors"
	"github/heimaolst/collectionbox/internal/biz"
	"github/heimaolst/collectionbox/internal/logx"
	"net/http"
	"time"
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

// --- 您的 Handler (重构后) ---
func (s *CollectionService) CreateCollection(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer r.Body.Close()
	}

	// 1. 解析 JSON (DTO)
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// 这是客户端错误 (400)
		writeError(w, http.StatusBadRequest, "Invalid JSON format: "+err.Error())
		return
	}

	// (可选的格式校验)
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}


	ctx := r.Context()
	cols, err := s.uc.CreateCollectionsFromText(ctx, req.URL)

	// 4. 【关键】翻译 biz 层错误
	if err != nil {
	
		if errors.Is(err, biz.ErrInvalidArgument) {
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			logx.FromContext(ctx).Error("internal error creating collections", "err", err)
			writeError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// 5. 返回成功响应（批量创建）
	writeJSON(w, http.StatusCreated, cols)
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
	//clean data
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
