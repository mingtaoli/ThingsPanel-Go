package services

import (
	"ThingsPanel-Go/initialize/psql"
	"ThingsPanel-Go/models"
	valid "ThingsPanel-Go/validate"

	"github.com/beego/beego/v2/core/logs"
	"gorm.io/gorm"
)

type TpOtaDeviceService struct {
	//可搜索字段
	SearchField []string
	//可作为条件的字段
	WhereField []string
	//可做为时间范围查询的字段
	TimeField []string
}
type DeviceStatusCount struct {
	UpgradeStatus string `json:"upgrade_status,omitempty" alias:"状态"`
	Count         int    `json:"count" alias:"数量"`
}

func (*TpOtaDeviceService) GetTpOtaDeviceList(PaginationValidate valid.TpOtaDevicePaginationValidate) (bool, []models.TpOtaDevice, int64) {
	var TpOtaDevices []models.TpOtaDevice
	offset := (PaginationValidate.CurrentPage - 1) * PaginationValidate.PerPage
	db := psql.Mydb.Model(&models.TpOtaDevice{})
	if PaginationValidate.OtaTaskId != "" {
		db.Where("ota_task_id =?", PaginationValidate.OtaTaskId)
	}
	if PaginationValidate.DeviceId != "" {
		db.Where("device_id like ?", "%"+PaginationValidate.DeviceId+"%")
	}
	if PaginationValidate.UpgradeStatus != "" {
		db.Where("upgrade_status =?", PaginationValidate.UpgradeStatus)
	}
	var count int64
	db.Count(&count)
	result := db.Limit(PaginationValidate.PerPage).Offset(offset).Find(&TpOtaDevices)
	if result.Error != nil {
		logs.Error(result.Error, gorm.ErrRecordNotFound)
		return false, TpOtaDevices, 0
	}
	return true, TpOtaDevices, count
}

func (*TpOtaDeviceService) GetTpOtaDeviceStatusCount(PaginationValidate valid.TpOtaDevicePaginationValidate) (bool, []DeviceStatusCount) {
	StatusCount := make([]DeviceStatusCount, 0)
	db := psql.Mydb.Model(&models.TpOtaDevice{})
	re := db.Select("upgrade_status as upgrade_status,count(*) as count").Where("remark = ? ", "ccc").Group("upgrade_status").Scan(&StatusCount)
	if re.Error != nil {
		return false, StatusCount
	}
	return true, StatusCount

}

// 新增数据
func (*TpOtaDeviceService) AddTpOtaDevice(tp_ota_device models.TpOtaDevice) (models.TpOtaDevice, error) {
	result := psql.Mydb.Create(&tp_ota_device)
	if result.Error != nil {
		logs.Error(result.Error, gorm.ErrRecordNotFound)
		return tp_ota_device, result.Error
	}
	return tp_ota_device, nil
}

//批量插入数据
func (*TpOtaDeviceService) AddBathTpOtaDevice(tp_ota_device []models.TpOtaDevice) ([]models.TpOtaDevice, error) {
	result := psql.Mydb.Create(&tp_ota_device)
	if result.Error != nil {
		logs.Error(result.Error, gorm.ErrRecordNotFound)
		return tp_ota_device, result.Error
	}
	return tp_ota_device, nil
}

//修改升级状态
//0-待推送 1-已推送 2-升级中 修改为已取消
//4-升级失败 修改为待推送
//3-升级成功 5-已取消 不修改
func (*TpOtaDeviceService) ModfiyUpdateDevice(tp_ota_device models.TpOtaDevice) error {
	var upgrade_status_list []string
	var result *gorm.DB
	db := psql.Mydb.Model(&models.TpOtaDevice{})
	if tp_ota_device.OtaTaskId != "" {
		result = db.Select("upgrade_status").Where("ota_task_id=?", tp_ota_device.OtaTaskId).Find(&upgrade_status_list)

	} else {
		result = db.Select("upgrade_status").Where("id=?", tp_ota_device.Id).Find(&upgrade_status_list)
	}
	if result.Error != nil {
		logs.Error(result.Error)
		return result.Error
	}
	for _, upgrade_status := range upgrade_status_list {
		if upgrade_status == "0" || upgrade_status == "1" || upgrade_status == "2" {
			db.Where("id=?", tp_ota_device.Id).Update("upgrade_status", "5")
		}
		if upgrade_status == "4" {
			db.Where("id=?", tp_ota_device.Id).Update("upgrade_status", "0")
		}
	}
	return nil
}
