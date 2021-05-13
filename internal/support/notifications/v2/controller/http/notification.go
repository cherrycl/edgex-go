//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"math"
	"net/http"
	"strconv"

	"github.com/edgexfoundry/edgex-go/internal/pkg"
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation"
	"github.com/edgexfoundry/edgex-go/internal/pkg/v2/utils"
	notificationContainer "github.com/edgexfoundry/edgex-go/internal/support/notifications/container"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/application"
	"github.com/edgexfoundry/edgex-go/internal/support/notifications/v2/io"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/common"
	requestDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/requests"
	responseDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/v2/dtos/responses"

	"github.com/gorilla/mux"
)

type NotificationController struct {
	reader io.NotificationReader
	dic    *di.Container
}

// NewNotificationController creates and initializes an NotificationController
func NewNotificationController(dic *di.Container) *NotificationController {
	return &NotificationController{
		reader: io.NewNotificationRequestReader(),
		dic:    dic,
	}
}

func (nc *NotificationController) AddNotification(w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	lc := container.LoggingClientFrom(nc.dic.Get)

	ctx := r.Context()
	correlationId := correlation.FromContext(ctx)

	addNotificationDTOs, err := nc.reader.ReadAddNotificationRequest(r.Body)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	notifications := requestDTO.AddNotificationReqToNotificationModels(addNotificationDTOs)

	var addResponses []interface{}
	for i, n := range notifications {
		var response interface{}
		reqId := addNotificationDTOs[i].RequestId
		newId, err := application.AddNotification(n, ctx, nc.dic)
		if err != nil {
			lc.Error(err.Error(), clients.CorrelationHeader, correlationId)
			lc.Debug(err.DebugMessages(), clients.CorrelationHeader, correlationId)
			response = commonDTO.NewBaseResponse(
				reqId,
				err.Message(),
				err.Code())
		} else {
			response = commonDTO.NewBaseWithIdResponse(
				reqId,
				"",
				http.StatusCreated,
				newId)
		}
		addResponses = append(addResponses, response)
	}

	utils.WriteHttpHeader(w, ctx, http.StatusMultiStatus)
	pkg.Encode(addResponses, w, lc)
}

func (nc *NotificationController) NotificationById(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(nc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	id := vars[v2.Id]

	notification, err := application.NotificationById(id, nc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewNotificationResponse("", "", http.StatusOK, notification)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (nc *NotificationController) NotificationsByCategory(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(nc.dic.Get)
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(nc.dic.Get)

	vars := mux.Vars(r)
	category := vars[v2.Category]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	notifications, err := application.NotificationsByCategory(offset, limit, category, nc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiNotificationsResponse("", "", http.StatusOK, notifications)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (nc *NotificationController) NotificationsByLabel(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(nc.dic.Get)
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(nc.dic.Get)

	vars := mux.Vars(r)
	label := vars[v2.Label]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	notifications, err := application.NotificationsByLabel(offset, limit, label, nc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiNotificationsResponse("", "", http.StatusOK, notifications)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (nc *NotificationController) NotificationsByStatus(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(nc.dic.Get)
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(nc.dic.Get)

	vars := mux.Vars(r)
	status := vars[v2.Status]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	notifications, err := application.NotificationsByStatus(offset, limit, status, nc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiNotificationsResponse("", "", http.StatusOK, notifications)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (nc *NotificationController) NotificationsByTimeRange(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(nc.dic.Get)
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(nc.dic.Get)

	// parse time range (start, end), offset, and limit from incoming request
	start, end, offset, limit, err := utils.ParseTimeRangeOffsetLimit(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	notifications, err := application.NotificationsByTimeRange(start, end, offset, limit, nc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiNotificationsResponse("", "", http.StatusOK, notifications)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

func (nc *NotificationController) DeleteNotificationById(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(nc.dic.Get)
	ctx := r.Context()

	// URL parameters
	vars := mux.Vars(r)
	id := vars[v2.Id]

	err := application.DeleteNotificationById(id, nc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusOK)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

// NotificationsBySubscriptionName queries notifications by offset, limit and subscriptionName
func (nc *NotificationController) NotificationsBySubscriptionName(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(nc.dic.Get)
	ctx := r.Context()
	config := notificationContainer.ConfigurationFrom(nc.dic.Get)

	vars := mux.Vars(r)
	subscriptionName := vars[v2.Name]

	// parse URL query string for offset, limit
	offset, limit, _, err := utils.ParseGetAllObjectsRequestQueryString(r, 0, math.MaxInt32, -1, config.Service.MaxResultCount)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	notifications, err := application.NotificationsBySubscriptionName(offset, limit, subscriptionName, nc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := responseDTO.NewMultiNotificationsResponse("", "", http.StatusOK, notifications)
	utils.WriteHttpHeader(w, ctx, http.StatusOK)
	pkg.Encode(response, w, lc)
}

// CleanupNotificationsByAge deletes notifications which have age and is less than the specified one, where the age of Notification is calculated by subtracting its last modification timestamp from the current timestamp. Note that the corresponding transmissions will also be deleted.
func (nc *NotificationController) CleanupNotificationsByAge(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(nc.dic.Get)
	ctx := r.Context()

	vars := mux.Vars(r)
	age, parsingErr := strconv.ParseInt(vars[v2.Age], 10, 64)
	if parsingErr != nil {
		err := errors.NewCommonEdgeX(errors.KindContractInvalid, "age format parsing failed", parsingErr)
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}
	err := application.CleanupNotificationsByAge(age, nc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusAccepted)
	utils.WriteHttpHeader(w, ctx, http.StatusAccepted)
	// encode and send out the response
	pkg.Encode(response, w, lc)
}

// CleanupNotifications deletes all notifications and the corresponding transmissions.
func (nc *NotificationController) CleanupNotifications(w http.ResponseWriter, r *http.Request) {
	lc := container.LoggingClientFrom(nc.dic.Get)
	ctx := r.Context()

	// Use zero as the age to delete all
	err := application.CleanupNotificationsByAge(0, nc.dic)
	if err != nil {
		utils.WriteErrorResponse(w, ctx, lc, err, "")
		return
	}

	response := commonDTO.NewBaseResponse("", "", http.StatusAccepted)
	utils.WriteHttpHeader(w, ctx, http.StatusAccepted)
	// encode and send out the response
	pkg.Encode(response, w, lc)
}