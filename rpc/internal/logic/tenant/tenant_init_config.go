package tenant

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coder-lulu/newbee-core/rpc/ent"
)

// InitializationTemplate 初始化模板配置
type InitializationTemplate struct {
	Name        string                `json:"name"`
	Version     string                `json:"version"`
	Description string                `json:"description"`
	Components  []ComponentTemplate   `json:"components"`
}

// ComponentTemplate 组件初始化模板
type ComponentTemplate struct {
	Name         string      `json:"name"`
	Dependencies []string    `json:"dependencies"`
	Data         interface{} `json:"data"`
	Required     bool        `json:"required"`
}

// DictionaryTemplate 字典模板
type DictionaryTemplate struct {
	Title   string                   `json:"title"`
	Name    string                   `json:"name"`
	Desc    string                   `json:"desc"`
	Details []DictionaryDetailTemplate `json:"details"`
}

// DictionaryDetailTemplate 字典详情模板
type DictionaryDetailTemplate struct {
	Title     string `json:"title"`
	Value     string `json:"value"`
	Sort      int    `json:"sort"`
	IsDefault uint32 `json:"is_default"`
	ListClass string `json:"list_class"`
	CSSClass  string `json:"css_class"`
}

// ConfigurationTemplate 配置模板
type ConfigurationTemplate struct {
	Name     string `json:"name"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	Category string `json:"category"`
	Remark   string `json:"remark"`
	Sort     int    `json:"sort"`
}

// PositionTemplate 职位模板
type PositionTemplate struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Remark string `json:"remark"`
	Sort   int    `json:"sort"`
}

// APIPermissionTemplate API权限模板
type APIPermissionTemplate struct {
	Path        string `json:"path"`
	Description string `json:"description"`
	APIGroup    string `json:"api_group"`
	Method      string `json:"method"`
}

// GetDefaultInitTemplate 获取默认初始化模板
func GetDefaultInitTemplate() *InitializationTemplate {
	return &InitializationTemplate{
		Name:        "标准企业模板",
		Version:     "1.0.0",
		Description: "标准企业租户初始化模板，包含基础的字典、配置、权限等数据",
		Components: []ComponentTemplate{
			{
				Name:         "dictionaries",
				Dependencies: []string{},
				Required:     true,
				Data: []DictionaryTemplate{
					{
						Title: "用户性别",
						Name:  "sys_user_sex",
						Desc:  "用户性别列表",
						Details: []DictionaryDetailTemplate{
							{Title: "男", Value: "0", Sort: 1, IsDefault: 1, ListClass: "default"},
							{Title: "女", Value: "1", Sort: 2, IsDefault: 0, ListClass: "default"},
							{Title: "未知", Value: "2", Sort: 3, IsDefault: 0, ListClass: "default"},
						},
					},
					{
						Title: "系统状态",
						Name:  "sys_common_status",
						Desc:  "通用状态列表",
						Details: []DictionaryDetailTemplate{
							{Title: "正常", Value: "1", Sort: 1, IsDefault: 1, ListClass: "success"},
							{Title: "停用", Value: "0", Sort: 2, IsDefault: 0, ListClass: "danger"},
						},
					},
					{
						Title: "系统是否",
						Name:  "sys_yes_no",
						Desc:  "系统是否列表",
						Details: []DictionaryDetailTemplate{
							{Title: "是", Value: "1", Sort: 1, IsDefault: 0, ListClass: "primary"},
							{Title: "否", Value: "0", Sort: 2, IsDefault: 1, ListClass: "default"},
						},
					},
				},
			},
			{
				Name:         "configurations",
				Dependencies: []string{},
				Required:     true,
				Data: []ConfigurationTemplate{
					{Name: "系统名称", Key: "system.title", Value: "NewBee管理系统", Category: "system", Remark: "系统页面标题", Sort: 1},
					{Name: "用户初始密码", Key: "system.user.init_password", Value: "123456", Category: "system", Remark: "新用户的初始密码", Sort: 2},
					{Name: "会话超时时间", Key: "system.session.timeout", Value: "30", Category: "system", Remark: "用户会话超时时间(分钟)", Sort: 3},
					{Name: "密码复杂度检查", Key: "system.password.complexity_check", Value: "true", Category: "security", Remark: "是否启用密码复杂度检查", Sort: 4},
					{Name: "登录失败锁定", Key: "system.login.lock_enabled", Value: "true", Category: "security", Remark: "是否启用登录失败锁定", Sort: 5},
					{Name: "最大登录失败次数", Key: "system.login.max_retry_count", Value: "5", Category: "security", Remark: "最大登录失败次数", Sort: 6},
				},
			},
			{
				Name:         "positions",
				Dependencies: []string{"departments"},
				Required:     true,
				Data: []PositionTemplate{
					{Name: "总经理", Code: "general_manager", Remark: "企业最高管理者", Sort: 1},
					{Name: "部门经理", Code: "department_manager", Remark: "部门负责人", Sort: 2},
					{Name: "项目经理", Code: "project_manager", Remark: "项目负责人", Sort: 3},
					{Name: "高级工程师", Code: "senior_engineer", Remark: "高级技术人员", Sort: 4},
					{Name: "普通员工", Code: "employee", Remark: "普通工作人员", Sort: 5},
				},
			},
			{
				Name:         "api_permissions",
				Dependencies: []string{},
				Required:     false,
				Data: []APIPermissionTemplate{
					// 用户管理
					{Path: "/api/v1/user/list", Description: "获取用户列表", APIGroup: "user", Method: "GET"},
					{Path: "/api/v1/user/create", Description: "创建用户", APIGroup: "user", Method: "POST"},
					{Path: "/api/v1/user/update", Description: "更新用户", APIGroup: "user", Method: "PUT"},
					{Path: "/api/v1/user/delete", Description: "删除用户", APIGroup: "user", Method: "DELETE"},
					// 角色管理
					{Path: "/api/v1/role/list", Description: "获取角色列表", APIGroup: "role", Method: "GET"},
					{Path: "/api/v1/role/create", Description: "创建角色", APIGroup: "role", Method: "POST"},
					{Path: "/api/v1/role/update", Description: "更新角色", APIGroup: "role", Method: "PUT"},
					{Path: "/api/v1/role/delete", Description: "删除角色", APIGroup: "role", Method: "DELETE"},
					// 部门管理
					{Path: "/api/v1/department/list", Description: "获取部门列表", APIGroup: "department", Method: "GET"},
					{Path: "/api/v1/department/create", Description: "创建部门", APIGroup: "department", Method: "POST"},
					{Path: "/api/v1/department/update", Description: "更新部门", APIGroup: "department", Method: "PUT"},
					{Path: "/api/v1/department/delete", Description: "删除部门", APIGroup: "department", Method: "DELETE"},
				},
			},
		},
	}
}

// TenantInitializer 租户初始化器接口
type TenantInitializer interface {
	Initialize(ctx context.Context, tx *ent.Tx, tenantID uint64, template *InitializationTemplate) error
	Rollback(ctx context.Context, tx *ent.Tx, tenantID uint64) error
	Validate(template *InitializationTemplate) error
}

// DefaultTenantInitializer 默认租户初始化器
type DefaultTenantInitializer struct {
	logger interface{ Infow(msg string, keysAndValues ...interface{}) }
}

// NewDefaultTenantInitializer 创建默认租户初始化器
func NewDefaultTenantInitializer(logger interface{ Infow(msg string, keysAndValues ...interface{}) }) *DefaultTenantInitializer {
	return &DefaultTenantInitializer{logger: logger}
}

// Validate 验证初始化模板
func (d *DefaultTenantInitializer) Validate(template *InitializationTemplate) error {
	if template.Name == "" {
		return ErrInvalidTemplate{Reason: "模板名称不能为空"}
	}
	if template.Version == "" {
		return ErrInvalidTemplate{Reason: "模板版本不能为空"}
	}
	if len(template.Components) == 0 {
		return ErrInvalidTemplate{Reason: "至少需要一个组件"}
	}

	// 验证组件依赖关系
	componentNames := make(map[string]bool)
	for _, comp := range template.Components {
		componentNames[comp.Name] = true
	}

	for _, comp := range template.Components {
		for _, dep := range comp.Dependencies {
			if !componentNames[dep] {
				return ErrInvalidTemplate{
					Reason: fmt.Sprintf("组件 %s 依赖的组件 %s 不存在", comp.Name, dep),
				}
			}
		}
	}

	return nil
}

// ErrInvalidTemplate 无效模板错误
type ErrInvalidTemplate struct {
	Reason string
}

func (e ErrInvalidTemplate) Error() string {
	return fmt.Sprintf("无效的初始化模板: %s", e.Reason)
}

// InitConfigHelper 初始化配置助手
type InitConfigHelper struct{}

// LoadTemplate 加载初始化模板
func (h *InitConfigHelper) LoadTemplate(data []byte) (*InitializationTemplate, error) {
	var template InitializationTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("解析模板失败: %w", err)
	}
	return &template, nil
}

// SaveTemplate 保存初始化模板
func (h *InitConfigHelper) SaveTemplate(template *InitializationTemplate) ([]byte, error) {
	return json.MarshalIndent(template, "", "  ")
}

// MergeTemplates 合并多个模板
func (h *InitConfigHelper) MergeTemplates(base *InitializationTemplate, overlays ...*InitializationTemplate) *InitializationTemplate {
	result := *base
	
	for _, overlay := range overlays {
		// 合并组件
		componentMap := make(map[string]ComponentTemplate)
		for _, comp := range result.Components {
			componentMap[comp.Name] = comp
		}
		
		for _, comp := range overlay.Components {
			componentMap[comp.Name] = comp
		}
		
		result.Components = make([]ComponentTemplate, 0, len(componentMap))
		for _, comp := range componentMap {
			result.Components = append(result.Components, comp)
		}
	}
	
	return &result
}