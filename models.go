package main

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	ID        uint64    `gorm:"primaryKey; autoIncrement; not_null;"`
	Name      string    `gorm:"uniqueIndex; index; type:mediumtext not null"`
	Price     float64   `gorm:"not null;"`
	Type      string    `gorm:"not null; default:'food'"`
	CreatedAt time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time `gorm:"autoCreateTime:milli"`
}

type Order struct {
	gorm.Model
	ID            uint64         `gorm:"primaryKey; autoIncrement; not_null;"`
	Cancelled     uint8          `gorm:"not_null; default 0"`
	CreatedAt     time.Time      `gorm:"autoCreateTime:milli"`
	OrderProducts []OrderProduct `gorm:"foreignKey:OrderID;references:ID;"`
	Products      []Product      `gorm:"-"`
}

type OrderProduct struct {
	gorm.Model
	ID        uint64    `gorm:"primaryKey; autoIncrement; not_null;"`
	ProductID uint64    `gorm:"not_null;"` // index;
	OrderID   uint64    `gorm:"not_null;"` // index;
	CreatedAt time.Time `gorm:"autoCreateTime:milli"`
	Product   Product
}

type User struct {
	gorm.Model
	ID        uint64    `gorm:"primaryKey; autoIncrement; not_null;"`
	Username  string    `gorm:"uniqueIndex; index; not_null;"`
	Password  string    `gorm:"not_null;"`
	CreatedAt time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time `gorm:"autoCreateTime:milli"`
}

// https://gorm.io/docs/has_one.html#Override-Foreign-Key
