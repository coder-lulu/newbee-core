package tasklog

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/middleware/keys"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/suyuan32/simple-admin-job/jobclient"
	"github.com/suyuan32/simple-admin-job/types/job"
	"google.golang.org/grpc"
)

type recordingJob struct {
	jobclient.Job
	called  bool
	ctx     context.Context
	request *job.TaskLogListReq
}

func (j *recordingJob) GetTaskLogList(ctx context.Context, req *job.TaskLogListReq, _ ...grpc.CallOption) (*job.TaskLogListResp, error) {
	j.called = true
	j.ctx = ctx
	j.request = req
	return &job.TaskLogListResp{Total: 30, Data: []*job.TaskLogInfo{}}, nil
}

func TestTaskLogListOptionalFilters(t *testing.T) {
	for _, test := range []struct {
		name, body         string
		taskID             uint64
		result             uint32
		hasTask, hasResult bool
	}{
		{"task page without result", `{"page":1,"pageSize":20,"taskId":12}`, 12, 0, true, false},
		{"unfiltered page", `{"page":1,"pageSize":20}`, 0, 0, false, false},
		{"explicit filters", `{"page":1,"pageSize":20,"taskId":12,"result":2}`, 12, 2, true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			fake := &recordingJob{}
			s := &svc.ServiceContext{JobRpc: fake, Trans: i18n.NewTranslator(i18n.Conf{}, i18n.LocaleFS)}
			s.Config.JobRpc.Enabled = true
			ctx := keys.NewContextManager().SetTenantID(context.Background(), "1")
			ctx = context.WithValue(ctx, "lang", "zh_CN")
			req := httptest.NewRequest("POST", "/task_log/list", strings.NewReader(test.body)).WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			GetTaskLogListHandler(s)(response, req)
			if !fake.called {
				t.Fatalf("request rejected before RPC: HTTP %d", response.Code)
			}
			if fake.ctx != ctx || fake.request.Page != 1 || fake.request.PageSize != 20 {
				t.Fatal("request context or pagination lost")
			}
			if (fake.request.TaskId != nil) != test.hasTask || (fake.request.Result != nil) != test.hasResult || fake.request.GetTaskId() != test.taskID || fake.request.GetResult() != test.result {
				t.Fatal("optional filters not preserved")
			}
			var envelope struct {
				Code int `json:"code"`
				Data struct {
					Total int `json:"total"`
				} `json:"data"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			if envelope.Code != 0 || envelope.Data.Total != 30 {
				t.Fatal("list response envelope lost")
			}
		})
	}
}
