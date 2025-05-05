package handler

import (
	common "MoneyHook/MoneyHook-API/common"
	"MoneyHook/MoneyHook-API/model"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func (h *Handler) ProcessDailyJob(c echo.Context) error {
	error := validHeaders(c)
	if error != nil {
		fmt.Println(error)
		return c.JSON(http.StatusForbidden, error)
	}

	fixed_list := selectMonthlyTransactions(h)

	if len(*fixed_list) == 0 {
		return c.String(http.StatusOK, "Today is Nothing, Success Jobs")
	}

	log.Println("=== Start  InsertTransaction ===")
	var transactions []model.JobTransaction
	for _, fixed := range *fixed_list {
		transaction := model.JobTransaction{
			UserNo:            fixed.UserNo,
			TransactionName:   fixed.MonthlyTransactionName,
			TransactionAmount: fixed.MonthlyTransactionAmount,
			TransactionDate:   time.Now(),
			CategoryId:        fixed.CategoryId,
			SubCategoryId:     fixed.SubCategoryId,
			FixedFlg:          true,
			PaymentId:         fixed.PaymentId,
		}
		transactions = append(transactions, transaction)
	}

	err := h.jobsStore.InsertTransaction(&transactions)
	if err != nil {
		log.Printf("=== Failed InsertTransaction: %v ===\n", err)
		message := "Failed to insert transaction"
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(&message))
	}
	log.Println("=== Finish InsertTransaction ===")

	return c.String(http.StatusOK, "Success Jobs")
}

func validHeaders(c echo.Context) map[string]string {
	user_agent := c.Request().Header.Get(model.UserAgent)
	content_type := c.Request().Header.Get(model.ContentType)
	x_cloud_scheduler, err := strconv.ParseBool(c.Request().Header.Get(model.XCloudScheduler))
	x_cloud_scheduler_job_name := c.Request().Header.Get(model.XCloudSchedulerJobName)
	x_cloud_scheduler_schedule_time := c.Request().Header.Get(model.XCloudSchedulerScheduleTime)
	invalidRequest := "Invalid Request"

	fmt.Println("user_agent:", user_agent)
	fmt.Println("content_type:", content_type)
	fmt.Println("x_cloud_scheduler:", x_cloud_scheduler)
	fmt.Println("x_cloud_scheduler_job_name:", x_cloud_scheduler_job_name)
	fmt.Println("x_cloud_scheduler_schedule_time:", x_cloud_scheduler_schedule_time)

	switch {
	case user_agent != "Google-Cloud-Scheduler":
		log.Printf("Invalid User-Agent: '%s'", user_agent)
		return model.Error.Create(&invalidRequest)
	case content_type != "application/octet-stream":
		log.Printf("Invalid Content-Type: '%s'", content_type)
		return model.Error.Create(&invalidRequest)
	case !x_cloud_scheduler || err != nil:
		log.Printf("Invalid X-CloudScheduler: '%s'", c.Request().Header.Get(model.XCloudScheduler))
		return model.Error.Create(&invalidRequest)
	case x_cloud_scheduler_job_name != common.GetEnv("JOB_NAME", ""):
		log.Printf("Invalid X-CloudScheduler-JobName: '%s'", x_cloud_scheduler_job_name)
		return model.Error.Create(&invalidRequest)
	case x_cloud_scheduler_schedule_time == "":
		log.Printf("Invalid X-CloudScheduler-ScheduleTime: '%s'", x_cloud_scheduler_schedule_time)
		return model.Error.Create(&invalidRequest)
	}
	return nil
}

func selectMonthlyTransactions(h *Handler) *[]model.JobMonthlyTransaction {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	today := time.Now().In(jst)
	log.Printf("Today is %s\n", today.Format("2006-01-02"))

	year, month, day := today.Year(), today.Month(), today.Day()
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	firstOfNextMonth := firstOfMonth.AddDate(0, 1, 0)
	lastOfMonth := firstOfNextMonth.AddDate(0, 0, -1)
	lastDay := lastOfMonth.Day()
	log.Printf("Last day of this month: %d\n", lastDay)

	return h.jobsStore.SelectMonthlyTransaction(day, lastDay == day)
}
