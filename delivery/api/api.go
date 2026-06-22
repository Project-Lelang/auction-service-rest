package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	_ "auction-service/docs"

	"auction-service/constant"
	"auction-service/delivery/dto_response"
	"auction-service/delivery/middleware"
	"auction-service/delivery/ws"
	"auction-service/global"
	internalValidator "auction-service/internal/validator"
	"auction-service/manager"
	"auction-service/model"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type apiContext struct {
	ginCtx *gin.Context
}

func newApiContext(ctx *gin.Context) apiContext {
	return apiContext{ginCtx: ctx}
}

func (a *apiContext) context() context.Context {
	return a.ginCtx.Request.Context()
}

func (a *apiContext) getParam(key string) string {
	return a.ginCtx.Param(key)
}

func (a *apiContext) mustGetParamInt64(key string) int64 {
	id, err := strconv.ParseInt(a.getParam(key), 10, 64)
	if err != nil || id <= 0 {
		panic(dto_response.NewBadRequestErrorResponse(constant.LanguageSystemInvalidRequestPayload))
	}
	return id
}

func (a *apiContext) getQuery(key string) string {
	return a.ginCtx.Query(key)
}

func (a *apiContext) getQueryDefault(key, def string) string {
	return a.ginCtx.DefaultQuery(key, def)
}

func (a *apiContext) shouldBind(obj interface{}) error {
	err := a.ginCtx.ShouldBindJSON(obj)
	if errors.Is(err, io.EOF) {
		// Empty body is acceptable; all fields keep their zero / nil values.
		return nil
	}
	return err
}

func (a *apiContext) mustBind(obj interface{}) {
	if err := a.shouldBind(obj); err != nil {
		panic(a.translateBindErr(err))
	}
}

func (a *apiContext) shouldBindQuery(obj interface{}) error {
	return a.ginCtx.ShouldBindQuery(obj)
}

func (a *apiContext) mustBindQuery(obj interface{}) {
	if err := a.shouldBindQuery(obj); err != nil {
		panic(a.translateBindErr(err))
	}
}

func (a *apiContext) translateBindErr(err error) dto_response.ErrorResponse {
	switch v := err.(type) {
	case validator.ValidationErrors:
		errs := []dto_response.Error{}

		trans, ok := model.GetValidatorTranslatorCtx(a.context())
		if ok {
			translations := v.Translate(trans)
			for field, msg := range translations {
				errs = append(errs, dto_response.Error{
					Domain:  field,
					Message: msg,
				})
			}
		} else {
			for _, fe := range v {
				errs = append(errs, dto_response.Error{
					Domain:  fe.Field(),
					Message: fe.Error(),
				})
			}
		}

		r := dto_response.NewBadRequestErrorResponse(constant.LanguageSystemInvalidRequestPayload)
		r.Errors = errs
		return r

	case *json.UnmarshalTypeError:
		return dto_response.NewBadRequestErrorResponse(constant.LanguageSystemInvalidRequestPayload)

	default:
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return dto_response.NewBadRequestErrorResponse(constant.LanguageSystemInvalidRequestPayload)
		}
		panic(err)
	}
}

func (a *apiContext) mustBindForm(obj interface{}) {
	if err := a.ginCtx.ShouldBindWith(obj, binding.FormMultipart); err != nil {
		panic(a.translateBindErr(err))
	}
}

func (a *apiContext) json(code int, obj interface{}) {
	a.ginCtx.JSON(code, obj)
}

func (a *apiContext) status(code int) {
	a.ginCtx.Status(code)
}

type api struct{}

func newApi() api {
	return api{}
}

func (a *api) Authorize(fn func(ctx apiContext)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		apiCtx := newApiContext(ctx)
		model.MustGetUserCtx(apiCtx.context())
		fn(apiCtx)
	}
}

// AuthorizeRoles requires the caller to have at least one of the specified roles.
// SUPERADMIN always passes.
func (a *api) AuthorizeRoles(roles []string, fn func(ctx apiContext)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		apiCtx := newApiContext(ctx)
		claims := model.MustGetUserCtx(apiCtx.context())
		for _, required := range roles {
			if claims.HasRole(required) {
				fn(apiCtx)
				return
			}
		}
		panic(constant.ErrForbidden)
	}
}

func (a *api) Guest(fn func(ctx apiContext)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fn(newApiContext(ctx))
	}
}
func registerMiddlewares(router gin.IRouter, container *manager.Container) {
	middleware.RequestIdHandler(router)
	middleware.TranslatorHandler(router)
	middleware.PanicHandler(router, container.InfrastructureManager().GetLoggerStack())
}

func registerRoutes(router *gin.Engine, container *manager.Container, hub *ws.Hub) {
	// init validator only once (triggers binding.Validator assignment)
	_ = internalValidator.Translators
	baseApi := newApi()

	wsHandler := ws.NewWsHandler(hub)
	router.GET("/ws/auctions/:auction_id", wsHandler.ServeWs)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiGroup := router.Group("/")
	middleware.JWTHandler(apiGroup, container.UseCaseManager().AuthUseCase())

	// admin
	RegisterAdminAuthApi(apiGroup, &baseApi, container.UseCaseManager())
	RegisterAdminUserApi(apiGroup, &baseApi, container.UseCaseManager())
	RegisterAdminProductApi(apiGroup, &baseApi, container.UseCaseManager())
	RegisterAdminRoleRequestApi(apiGroup, &baseApi, container.UseCaseManager())
	RegisterAdminWithdrawalRequestApi(apiGroup, &baseApi, container.UseCaseManager())
	RegisterAdminPaymentMethodApi(apiGroup, &baseApi, container.UseCaseManager())

	// user
	RegisterAuthApi(apiGroup, &baseApi, container.UseCaseManager())
	RegisterProductApi(apiGroup, &baseApi, container.UseCaseManager())
	RegisterOwnApi(apiGroup, &baseApi, container.UseCaseManager())
	RegisterUserRoleRequestApi(apiGroup, &baseApi, container.UseCaseManager())
	RegisterAuctionApi(apiGroup, &baseApi, container.UseCaseManager())
	RegisterUserAddressApi(apiGroup, &baseApi, container.UseCaseManager())
	RegisterBiteshipApi(apiGroup, &baseApi, container.UseCaseManager())

	// file
	RegisterFileApi(apiGroup, &baseApi, container.FilesystemManager(), container.BaseFileUseCase())
}

func NewRouter(container *manager.Container, hub *ws.Hub) *gin.Engine {
	if global.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(
		cors.New(cors.Config{
			AllowOrigins:     global.GetConfig().CorsAllowedOrigins,
			AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead},
			AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "Content-Length", "Origin", "Accept-Language"},
			ExposeHeaders:    []string{"Content-Type", "Content-Length", "X-Request-Id"},
			AllowCredentials: true,
			MaxAge:           2 * time.Hour,
		}),
	)

	registerMiddlewares(router, container)

	registerRoutes(router, container, hub)

	return router
}
