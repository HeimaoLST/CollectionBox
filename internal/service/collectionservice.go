package service

import (
	"encoding/json"
	"errors"
	"github/heimaolst/collectionbox/internal/biz"
	"log"
	"net/http"
)

type CollectionService struct {
	uc *biz.CollectionUsecase
}

type CreateRequest struct {
	URL string `json:"url"`
}
type GetRequest struct {
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

	// 3. 调用 biz 层
	// 【最佳实践】使用 r.Context() 而不是 context.Background()
	// 这样如果客户端断开连接，biz 层的操作可以被取消
	ctx := r.Context()
	col, err := s.uc.CreateCollection(ctx, req.URL)

	// 4. 【关键】翻译 biz 层错误
	if err != nil {
		// 在这里，我们检查 biz 层返回的是哪种错误
		if errors.Is(err, biz.ErrInvalidArgument) {
			// 业务逻辑说：参数无效 (400)
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			// 未知错误，这是服务器内部错误 (500)
			// 【重要】不要把 err.Error() 暴露给客户端
			// 我们应该记录详细日志
			log.Printf("Internal server error: %v", err)
			// 只返回一个通用的错误信息
			writeError(w, http.StatusInternalServerError, "Internal server error")
		}
		return // 不要忘记返回
	}

	// 5. 返回成功响应
	// 业务创建成功，应该返回 201 Created
	writeJSON(w, http.StatusCreated, col)
}

func (s *CollectionService) GetByOrigin(w http.ResponseWriter, r *http.Request) {
	targetOrigin := r.FormValue("origin")

	if targetOrigin != "" {
		
	}
}
func (s *CollectionService) GetByTimeRange(w http.ResponseWriter, r *http.Request) {

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
		log.Printf("Failed to encode response: %v", err)
	}
}
