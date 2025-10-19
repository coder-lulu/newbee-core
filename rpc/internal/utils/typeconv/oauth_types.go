package typeconv

import (
	"encoding/json"
	"time"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
)

// ConvertAuthStyle converts protobuf uint64 AuthStyle to ent int AuthStyle
func ConvertAuthStyle(pbAuthStyle *uint64) *int {
	if pbAuthStyle == nil {
		return nil
	}
	return pointy.GetPointer(int(*pbAuthStyle))
}

// ConvertAuthStyleFromEnt converts ent int AuthStyle to protobuf uint64 AuthStyle
func ConvertAuthStyleFromEnt(entAuthStyle int) *uint64 {
	return pointy.GetPointer(uint64(entAuthStyle))
}

// ConvertCacheTTL converts protobuf int32 CacheTTL to ent int CacheTTL
func ConvertCacheTTL(pbCacheTTL *int32) *int {
	if pbCacheTTL == nil {
		return nil
	}
	return pointy.GetPointer(int(*pbCacheTTL))
}

// ConvertCacheTTLFromEnt converts ent int CacheTTL to protobuf int32 CacheTTL
func ConvertCacheTTLFromEnt(entCacheTTL int) *int32 {
	return pointy.GetPointer(int32(entCacheTTL))
}

// ConvertCount converts protobuf int32 count to ent int count
func ConvertCount(pbCount *int32) *int {
	if pbCount == nil {
		return nil
	}
	return pointy.GetPointer(int(*pbCount))
}

// ConvertCountFromEnt converts ent int count to protobuf int32 count
func ConvertCountFromEnt(entCount int) *int32 {
	return pointy.GetPointer(int32(entCount))
}

// ConvertLastUsedAt converts protobuf int64 timestamp to ent time.Time
func ConvertLastUsedAt(pbTimestamp *int64) *time.Time {
	if pbTimestamp == nil {
		return nil
	}
	return pointy.GetPointer(time.Unix(*pbTimestamp/1000, (*pbTimestamp%1000)*1000000))
}

// ConvertLastUsedAtFromEnt converts ent time.Time to protobuf int64 timestamp
func ConvertLastUsedAtFromEnt(entTime time.Time) *int64 {
	// Check if time is zero value (default for unset optional fields)
	if entTime.IsZero() {
		return nil
	}
	return pointy.GetPointer(entTime.UnixMilli())
}

// ConvertStatus converts protobuf uint32 status to ent uint8 status
func ConvertStatus(pbStatus *uint32) *uint8 {
	if pbStatus == nil {
		return nil
	}
	return pointy.GetPointer(uint8(*pbStatus))
}

// ConvertStatusFromEnt converts ent uint8 status to protobuf uint32 status
func ConvertStatusFromEnt(entStatus uint8) *uint32 {
	return pointy.GetPointer(uint32(entStatus))
}

// ConvertExtraConfig converts protobuf JSON string to ent map
func ConvertExtraConfig(pbExtraConfig *string) *map[string]interface{} {
	if pbExtraConfig == nil || *pbExtraConfig == "" {
		return nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(*pbExtraConfig), &config); err != nil {
		return nil
	}
	return &config
}

// ConvertExtraConfigFromEnt converts ent map to protobuf JSON string
func ConvertExtraConfigFromEnt(entConfig map[string]interface{}) *string {
	if entConfig == nil {
		return nil
	}

	data, err := json.Marshal(entConfig)
	if err != nil {
		return pointy.GetPointer("")
	}
	return pointy.GetPointer(string(data))
}
