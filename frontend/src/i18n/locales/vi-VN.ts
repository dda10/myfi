const viVN: Record<string, string> = {
  // Navigation
  "nav.overview": "Tổng quan",
  "nav.portfolio": "Danh mục",
  "nav.markets": "Thị trường",
  "nav.filter": "Bộ lọc",
  "nav.allocation": "Phân bổ",
  "nav.settings": "Cài đặt",

  // Header
  "header.search_placeholder": "Tìm mã cổ phiếu (VD: FPT, VNM)...",
  "header.theme_light": "Chuyển sang chế độ sáng",
  "header.theme_dark": "Chuyển sang chế độ tối",
  "header.notifications": "Thông báo",
  "header.settings": "Cấu hình AI",

  // Watchlist
  "watchlist.title": "Danh sách theo dõi",
  "watchlist.add": "Thêm vào danh sách",
  "watchlist.remove": "Xóa khỏi danh sách",
  "watchlist.empty": "Chưa có mã nào trong danh sách theo dõi",

  // Portfolio
  "portfolio.title": "Danh mục đầu tư",
  "portfolio.nav": "Giá trị tài sản ròng",
  "portfolio.holdings": "Danh mục nắm giữ",
  "portfolio.transactions": "Lịch sử giao dịch",
  "portfolio.total_value": "Tổng giá trị",
  "portfolio.unrealized_pl": "Lãi/Lỗ chưa thực hiện",
  "portfolio.realized_pl": "Lãi/Lỗ đã thực hiện",
  "portfolio.daily_change": "Thay đổi trong ngày",

  // Asset types
  "asset.vn_stock": "Cổ phiếu VN",
  "asset.crypto": "Tiền mã hóa",
  "asset.gold": "Vàng",
  "asset.savings": "Tiết kiệm",
  "asset.bonds": "Trái phiếu",
  "asset.bank": "Tài khoản ngân hàng",

  // Financial terms
  "finance.pe": "P/E",
  "finance.pb": "P/B",
  "finance.roe": "ROE",
  "finance.roa": "ROA",
  "finance.eps": "EPS",
  "finance.market_cap": "Vốn hóa",
  "finance.volume": "Khối lượng",
  "finance.ev_ebitda": "EV/EBITDA",
  "finance.div_yield": "Tỷ suất cổ tức",
  "finance.debt_equity": "Nợ/Vốn chủ",
  "finance.revenue_growth": "Tăng trưởng doanh thu",
  "finance.profit_growth": "Tăng trưởng lợi nhuận",

  // Chart
  "chart.candlestick": "Nến",
  "chart.volume": "Khối lượng",
  "chart.indicators": "Chỉ báo",
  "chart.interval": "Khung thời gian",
  "chart.drawing_tools": "Công cụ vẽ",

  // Table headers
  "table.symbol": "Mã",
  "table.name": "Tên",
  "table.price": "Giá",
  "table.change": "Thay đổi",
  "table.change_pct": "% Thay đổi",
  "table.volume": "KL",
  "table.exchange": "Sàn",
  "table.sector": "Ngành",

  // Sectors
  "sector.VNIT": "Công nghệ",
  "sector.VNIND": "Công nghiệp",
  "sector.VNCONS": "Tiêu dùng",
  "sector.VNCOND": "Tiêu dùng thiết yếu",
  "sector.VNHEAL": "Y tế",
  "sector.VNENE": "Năng lượng",
  "sector.VNUTI": "Tiện ích",
  "sector.VNREAL": "Bất động sản",
  "sector.VNFIN": "Tài chính",
  "sector.VNMAT": "Vật liệu",

  // Buttons
  "btn.save": "Lưu",
  "btn.cancel": "Hủy",
  "btn.delete": "Xóa",
  "btn.edit": "Sửa",
  "btn.add": "Thêm",
  "btn.apply": "Áp dụng",
  "btn.reset": "Đặt lại",
  "btn.export": "Xuất",
  "btn.close": "Đóng",

  // Errors
  "error.generic": "Đã xảy ra lỗi. Vui lòng thử lại.",
  "error.network": "Lỗi kết nối mạng.",
  "error.not_found": "Không tìm thấy dữ liệu.",
  "error.insufficient_holdings": "Số lượng nắm giữ không đủ.",
  "error.invalid_asset_type": "Loại tài sản không hợp lệ.",

  // Validation
  "validation.required": "Trường này là bắt buộc",
  "validation.min_value": "Giá trị tối thiểu là {{value}}",
  "validation.max_value": "Giá trị tối đa là {{value}}",

  // Alerts
  "alert.price_alert_triggered": "Cảnh báo giá {{symbol}} đã kích hoạt",
  "alert.pattern_detected": "Phát hiện mẫu hình {{pattern}} cho {{symbol}}",

  // Dynamic
  "portfolio.value_label": "Giá trị danh mục: {{amount}}",
  "portfolio.change_label": "Thay đổi: {{amount}} ({{percent}})",

  // Screener
  "screener.title": "Bộ lọc cổ phiếu",
  "screener.filters": "Bộ lọc",
  "screener.results": "Kết quả",
  "screener.no_results": "Không tìm thấy kết quả phù hợp",
  "screener.save_preset": "Lưu bộ lọc",

  // Settings
  "settings.title": "Cài đặt",
  "settings.language": "Ngôn ngữ",
  "settings.theme": "Giao diện",
  "settings.ai_config": "Cấu hình AI",

  // Help
  "help.tooltip_pe": "Tỷ lệ giá trên thu nhập",
  "help.tooltip_pb": "Tỷ lệ giá trên giá trị sổ sách",
  "help.tooltip_roe": "Tỷ suất lợi nhuận trên vốn chủ sở hữu",

  // Comparison
  "comparison.title": "So sánh cổ phiếu",
  "comparison.add_symbol": "Thêm mã CK...",
  "comparison.clear_all": "Xóa tất cả",
  "comparison.valuation": "Định giá",
  "comparison.performance": "Hiệu suất",
  "comparison.correlation": "Tương quan",
  "comparison.select_prompt": "Chọn ít nhất 2 mã cổ phiếu để so sánh",
  "comparison.sector_dropdown": "Chọn ngành để thêm tự động",
  "comparison.pe_ratio": "P/E",
  "comparison.pb_ratio": "P/B",
  "comparison.return_pct": "Lợi nhuận %",
  "comparison.loading": "Đang tải dữ liệu...",
  "comparison.error": "Không thể tải dữ liệu so sánh",
  "comparison.max_stocks": "Tối đa 10 mã cổ phiếu",

  // Language selector
  "language.vi": "Tiếng Việt",
  "language.en": "English",
};

export default viVN;
