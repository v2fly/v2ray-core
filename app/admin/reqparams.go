// +build !confonly

package admin

type TableQuery struct {
	Search string `form:"search" `
	Sort   string `form:"sort"`
	Order  string `form:"order"`
	Limit  int32  `form:"limit"`
	Offset int32  `form:"offset"`
}
