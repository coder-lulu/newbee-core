package base

import (
	"context"

	"github.com/coder-lulu/newbee-common/orm/ent/hooks"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-core/rpc/ent"
)

// insert initial dict data
func (l *InitDatabaseLogic) insertDictData(ctx context.Context) error {
	tenantCtx := hooks.SetTenantIDToContext(context.Background(), 1)
	var dicts []*ent.DictionaryCreate
	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("用户性别").
		SetName("sys_user_sex").
		SetStatus(1).
		SetDesc("用户性别列表"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("菜单状态").
		SetName("sys_show_hide").
		SetStatus(1).
		SetDesc("菜单状态列表"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("系统开关").
		SetName("sys_normal_disable").
		SetStatus(1).
		SetDesc("系统开关列表"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("系统是否").
		SetName("sys_yes_no").
		SetStatus(1).
		SetDesc("系统是否列表"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("通知类型").
		SetName("sys_notice_type").
		SetStatus(1).
		SetDesc("通知类型列表"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("通知状态").
		SetName("sys_notice_status").
		SetStatus(1).
		SetDesc("通知状态列表"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("操作类型").
		SetName("sys_oper_type").
		SetStatus(1).
		SetDesc("操作类型列表"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("系统状态").
		SetName("sys_common_status").
		SetStatus(1).
		SetDesc("通用状态列表"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("授权类型").
		SetName("sys_grant_type").
		SetStatus(1).
		SetDesc("认证授权类型"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("设备类型").
		SetName("sys_device_type").
		SetStatus(1).
		SetDesc("客户端设备类型"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("业务状态").
		SetName("wf_business_status").
		SetStatus(1).
		SetDesc("业务状态列表"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("表单类型").
		SetName("wf_form_type").
		SetStatus(1).
		SetDesc("表单类型列表2"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("系统配置类型").
		SetName("sys_config_type").
		SetStatus(1).
		SetDesc("系统配置类型"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("系统菜单组件类型").
		SetName("sys_menu_component_type").
		SetStatus(1).
		SetDesc("系统菜单组件类型"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("系统RPC服务列表").
		SetName("sys_rpc_service_list").
		SetStatus(1).
		SetDesc("系统RPC服务列表"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("HTTP请求方法").
		SetName("sys_http_request_method").
		SetStatus(1).
		SetDesc("HTTP请求方法"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("系统文件类型").
		SetName("system_file_type").
		SetStatus(1).
		SetDesc("系统文件类型"),
	)

	// CMDB相关字典
	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("CI状态").
		SetName("cmdb_ci_status").
		SetStatus(1).
		SetDesc("配置项状态列表"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("属性类型").
		SetName("cmdb_attribute_type").
		SetStatus(1).
		SetDesc("CMDB属性类型列表"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("关系方向").
		SetName("cmdb_relation_direction").
		SetStatus(1).
		SetDesc("关系类型方向"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("CI生命周期状态").
		SetName("cmdb_ci_lifecycle_status").
		SetStatus(1).
		SetDesc("CI生命周期状态"),
	)

	dicts = append(dicts, l.svcCtx.DB.Dictionary.Create().
		SetTitle("权限类型").
		SetName("cmdb_permission_type").
		SetStatus(1).
		SetDesc("CMDB权限类型"),
	)

	err := l.svcCtx.DB.Dictionary.CreateBulk(dicts...).Exec(tenantCtx)
	if err != nil {
		logx.Errorw(err.Error())
		return errorx.NewInternalError(err.Error())
	}

	var dtypes []*ent.DictionaryDetailCreate
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("女").
		SetValue("1").
		SetDictionariesID(1).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("未知").
		SetValue("2").
		SetDictionariesID(1).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("男").
		SetValue("0").
		SetDictionariesID(1).
		SetCSSClass("").
		SetIsDefault(1).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("显示").
		SetValue("true").
		SetDictionariesID(2).
		SetCSSClass("dot-before-green").
		SetIsDefault(1).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("隐藏").
		SetValue("false").
		SetDictionariesID(2).
		SetCSSClass("dot-before-red").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("正常").
		SetValue("1").
		SetDictionariesID(3).
		SetCSSClass("dot-before-green").
		SetIsDefault(1).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("停用").
		SetValue("0").
		SetDictionariesID(3).
		SetCSSClass("dot-before-red").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("是").
		SetValue("1").
		SetDictionariesID(4).
		SetCSSClass("").
		SetIsDefault(1).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("否").
		SetValue("0").
		SetDictionariesID(4).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("公告").
		SetValue("2").
		SetDictionariesID(5).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("通知").
		SetValue("1").
		SetDictionariesID(5).
		SetCSSClass("").
		SetIsDefault(1).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("正常").
		SetValue("1").
		SetDictionariesID(6).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("关闭").
		SetValue("0").
		SetDictionariesID(6).
		SetCSSClass("").
		SetIsDefault(1).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("授权").
		SetValue("4").
		SetDictionariesID(7).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("其他").
		SetValue("0").
		SetDictionariesID(7).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("清空数据").
		SetValue("9").
		SetDictionariesID(7).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("生成代码").
		SetValue("8").
		SetDictionariesID(7).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("新增").
		SetValue("1").
		SetDictionariesID(7).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("强退").
		SetValue("7").
		SetDictionariesID(7).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("导入").
		SetValue("6").
		SetDictionariesID(7).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("导出").
		SetValue("5").
		SetDictionariesID(7).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("修改").
		SetValue("2").
		SetDictionariesID(7).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("删除").
		SetValue("3").
		SetDictionariesID(7).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("成功").
		SetValue("1").
		SetDictionariesID(8).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("失败").
		SetValue("0").
		SetDictionariesID(8).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("短信认证").
		SetValue("sms").
		SetDictionariesID(9).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("小程序认证").
		SetValue("xcx").
		SetDictionariesID(9).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("三方登录认证").
		SetValue("social").
		SetDictionariesID(9).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("密码认证").
		SetValue("password").
		SetDictionariesID(9).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("邮件认证").
		SetValue("email").
		SetDictionariesID(9).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("H5").
		SetValue("h5").
		SetDictionariesID(10).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("小程序").
		SetValue("xcx").
		SetDictionariesID(10).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("iOS").
		SetValue("ios").
		SetDictionariesID(10).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("安卓").
		SetValue("android").
		SetDictionariesID(10).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("PC").
		SetValue("pc").
		SetDictionariesID(10).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("已撤销").
		SetValue("cancel").
		SetDictionariesID(11).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("已作废").
		SetValue("invalid").
		SetDictionariesID(11).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("草稿").
		SetValue("draft").
		SetDictionariesID(11).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("已完成").
		SetValue("finished").
		SetDictionariesID(11).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("已终止").
		SetValue("termination").
		SetDictionariesID(11).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("已退回").
		SetValue("back").
		SetDictionariesID(11).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("待审核").
		SetValue("waiting").
		SetDictionariesID(11).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("自定义表单").
		SetValue("static").
		SetDictionariesID(12).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("动态表单").
		SetValue("dynamic").
		SetDictionariesID(12).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("内置").
		SetValue("internal").
		SetDictionariesID(13).
		SetCSSClass("").
		SetIsDefault(1).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("自定义").
		SetValue("custom").
		SetDictionariesID(13).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("cyan").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("内链").
		SetValue("InnerLink").
		SetDictionariesID(14).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("目录").
		SetValue("Layout").
		SetDictionariesID(14).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("外链").
		SetValue("ParentView").
		SetDictionariesID(14).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("核心服务").
		SetValue("Core").
		SetDictionariesID(15).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("定时任务").
		SetValue("Job").
		SetDictionariesID(15).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("文件管理").
		SetValue("Fms").
		SetDictionariesID(15).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("info").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("资产管理").
		SetValue("Cmdb").
		SetDictionariesID(15).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("info").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("PUT").
		SetValue("PUT").
		SetDictionariesID(16).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("orange").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("POST").
		SetValue("POST").
		SetDictionariesID(16).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("primary").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("DELETE").
		SetValue("DELETE").
		SetDictionariesID(16).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("red").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("GET").
		SetValue("GET").
		SetDictionariesID(16).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("purple").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("音频").
		SetValue("4").
		SetDictionariesID(17).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("其他").
		SetValue("1").
		SetDictionariesID(17).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("全部").
		SetValue("0").
		SetDictionariesID(17).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("图片").
		SetValue("2").
		SetDictionariesID(17).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("").
		SetStatus(1).
		SetSort(0),
	)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("视频").
		SetValue("3").
		SetDictionariesID(17).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("").
		SetStatus(1).
		SetSort(0),
	)

	// CMDB字典详情 - CI状态 (字典ID: 18)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("正常").
		SetValue("1").
		SetDictionariesID(18).
		SetCSSClass("dot-before-green").
		SetIsDefault(1).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("维护中").
		SetValue("2").
		SetDictionariesID(18).
		SetCSSClass("dot-before-orange").
		SetIsDefault(0).
		SetListClass("orange").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("故障").
		SetValue("3").
		SetDictionariesID(18).
		SetCSSClass("dot-before-red").
		SetIsDefault(0).
		SetListClass("red").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("下线").
		SetValue("4").
		SetDictionariesID(18).
		SetCSSClass("dot-before-gray").
		SetIsDefault(0).
		SetListClass("gray").
		SetStatus(1).
		SetSort(0),
	)

	// 属性类型 (字典ID: 19)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("文本").
		SetValue("text").
		SetDictionariesID(19).
		SetCSSClass("").
		SetIsDefault(1).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("整数").
		SetValue("integer").
		SetDictionariesID(19).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("primary").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("浮点数").
		SetValue("float").
		SetDictionariesID(19).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("primary").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("日期时间").
		SetValue("datetime").
		SetDictionariesID(19).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("info").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("JSON").
		SetValue("json").
		SetDictionariesID(19).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("purple").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("选择列表").
		SetValue("choice").
		SetDictionariesID(19).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("cyan").
		SetStatus(1).
		SetSort(0),
	)

	// 关系方向 (字典ID: 20)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("单向").
		SetValue("unidirectional").
		SetDictionariesID(20).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("default").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("双向").
		SetValue("bidirectional").
		SetDictionariesID(20).
		SetCSSClass("").
		SetIsDefault(1).
		SetListClass("primary").
		SetStatus(1).
		SetSort(0),
	)

	// CI生命周期状态 (字典ID: 21)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("规划中").
		SetValue("planning").
		SetDictionariesID(21).
		SetCSSClass("dot-before-blue").
		SetIsDefault(0).
		SetListClass("blue").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("开发中").
		SetValue("development").
		SetDictionariesID(21).
		SetCSSClass("dot-before-orange").
		SetIsDefault(0).
		SetListClass("orange").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("测试中").
		SetValue("testing").
		SetDictionariesID(21).
		SetCSSClass("dot-before-purple").
		SetIsDefault(0).
		SetListClass("purple").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("生产中").
		SetValue("production").
		SetDictionariesID(21).
		SetCSSClass("dot-before-green").
		SetIsDefault(1).
		SetListClass("green").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("已退役").
		SetValue("retired").
		SetDictionariesID(21).
		SetCSSClass("dot-before-gray").
		SetIsDefault(0).
		SetListClass("gray").
		SetStatus(1).
		SetSort(0),
	)

	// 权限类型 (字典ID: 22)
	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("只读").
		SetValue("read").
		SetDictionariesID(22).
		SetCSSClass("").
		SetIsDefault(1).
		SetListClass("info").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("读写").
		SetValue("write").
		SetDictionariesID(22).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("primary").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("管理").
		SetValue("admin").
		SetDictionariesID(22).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("red").
		SetStatus(1).
		SetSort(0),
	)

	dtypes = append(dtypes, l.svcCtx.DB.DictionaryDetail.Create().
		SetTitle("无权限").
		SetValue("none").
		SetDictionariesID(22).
		SetCSSClass("").
		SetIsDefault(0).
		SetListClass("gray").
		SetStatus(1).
		SetSort(0),
	)

	err = l.svcCtx.DB.DictionaryDetail.CreateBulk(dtypes...).Exec(tenantCtx)
	if err != nil {
		logx.Errorw(err.Error())
		return errorx.NewInternalError(err.Error())
	} else {
		return nil
	}
}
