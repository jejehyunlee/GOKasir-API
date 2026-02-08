package models

type BestSellingProduct struct {
	Name    string `json:"nama"`
	QtySold int    `json:"qty_terjual"`
}

type ReportResponse struct {
	TotalRevenue   int                `json:"total_revenue"`
	TotalTransaksi int                `json:"total_transaksi"`
	ProdukTerlaris BestSellingProduct `json:"produk_terlaris"`
}
