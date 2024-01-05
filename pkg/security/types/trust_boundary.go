/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/

package types

type TrustBoundary struct {
	Id                    string            `json:"id,omitempty"`
	Title                 string            `json:"title,omitempty"`
	Description           string            `json:"description,omitempty"`
	Type                  TrustBoundaryType `json:"type,omitempty"`
	Tags                  []string          `json:"tags,omitempty"`
	TechnicalAssetsInside []string          `json:"technical_assets_inside,omitempty"`
	TrustBoundariesNested []string          `json:"trust_boundaries_nested,omitempty"`
}

func (what TrustBoundary) IsTaggedWithAny(tags ...string) bool {
	return containsCaseInsensitiveAny(what.Tags, tags...)
}

func (what TrustBoundary) IsTaggedWithBaseTag(baseTag string) bool {
	return IsTaggedWithBaseTag(what.Tags, baseTag)
}

type ByTrustBoundaryTitleSort []TrustBoundary

func (what ByTrustBoundaryTitleSort) Len() int      { return len(what) }
func (what ByTrustBoundaryTitleSort) Swap(i, j int) { what[i], what[j] = what[j], what[i] }
func (what ByTrustBoundaryTitleSort) Less(i, j int) bool {
	return what[i].Title < what[j].Title
}
