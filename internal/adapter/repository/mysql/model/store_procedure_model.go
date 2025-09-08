package model

type DependencyCount struct {
	ParentID          uint `bun:"parent_id"`
	TotalDependencies int  `bun:"total_dependencies"`
}
