package main

import (
	"time"
)

type Ticket struct {
	// Ticket Identification
	TicketID   string `gorm:"primaryKey" json:"ticket_id"`
	TicketType string `json:"ticket_type"`
	Subject    string `json:"subject"`

	// Location & Organization
	Site        *string `json:"site,omitempty"`
	SiteGroup   *string `json:"site_group,omitempty"`
	RegionSite  *string `json:"region_site,omitempty"`
	DomainGroup string  `json:"domain_group"`
	Company     string  `json:"company"`
	Country     string  `json:"country"`

	// Status & Priority
	Priority     string `json:"priority"`
	TicketStatus string `json:"ticket_status"`
	CustomerName string `json:"customer_name"`

	// Product Categories
	ProductT1 *string // Identity Management
	ProductT2 *string // User Account Access
	ProductT3 *string // Onelogin

	// Assignment
	GroupAssignee *string // THL2 IT Security Identity Mgmt
	Assignee      *string // Pawaret

	// Timestamps
	CreatedDate  time.Time  // 06/07/2026 5:25 PM
	ResolvedAt   *time.Time // 07/07/2026 9:28 AM
	ClosedDate   *time.Time
	CompleteTime *time.Time

	// Details
	DetailDescription string

	// Source & Upload
	SourceFile string
	UploadedAt time.Time `gorm:"autoCreateTime" json:"uploaded_at"`
}
