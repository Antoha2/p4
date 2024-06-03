package helper

type Employee struct {
	Id            int    `gorm:"column:id"`
	Name          string `gorm:"column:name"`
	DepartamentId int    `gorm:"column:departament_id"`
	ProjectId     int    `gorm:"column:project_id"`
}

type Departament struct {
	Id      int    `gorm:"column:id"`
	DepName string `gorm:"column:dep_name"`
}
