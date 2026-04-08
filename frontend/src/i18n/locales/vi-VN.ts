const viVN: Record<string, string> = {
  // Navigation
  "nav.overview": "Tổng quan",
  "nav.dashboard": "Bảng điều khiển",
  "nav.portfolio": "Danh mục",
  "nav.markets": "Thị trường",
  "nav.filter": "Bộ lọc",
  "nav.allocation": "Phân bổ",
  "nav.settings": "Cài đặt",
  "nav.stock": "Cổ phiếu",
  "nav.screener": "Bộ lọc",
  "nav.ranking": "Xếp hạng",
  "nav.ideas": "Ý tưởng",
  "nav.heatmap": "Bản đồ nhiệt",
  "nav.missions": "Nhiệm vụ",
  "nav.research": "Nghiên cứu",
  "nav.macro": "Vĩ mô",
  "nav.chat": "Trò chuyện AI",

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
  "portfolio.buy": "Mua",
  "portfolio.sell": "Bán",
  "portfolio.cost_basis": "Giá vốn",
  "portfolio.quantity": "Số lượng",
  "portfolio.avg_price": "Giá trung bình",
  "portfolio.sector_allocation": "Phân bổ ngành",
  "portfolio.performance": "Hiệu suất",
  "portfolio.risk": "Rủi ro",
  "portfolio.sharpe": "Tỷ số Sharpe",
  "portfolio.max_drawdown": "Sụt giảm tối đa",
  "portfolio.beta": "Beta",
  "portfolio.volatility": "Biến động",

  // Asset types
  "asset.vn_stock": "Cổ phiếu VN",

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
  "chart.loading": "Đang tải biểu đồ…",
  "chart.error": "Không thể tải dữ liệu biểu đồ",
  "chart.no_data": "Không có dữ liệu",
  "chart.category.trend": "Xu hướng",
  "chart.category.momentum": "Động lượng",
  "chart.category.volatility": "Biến động",
  "chart.category.volume": "Khối lượng",
  "chart.drawing.trendline": "Đường xu hướng",
  "chart.drawing.horizontal": "Đường ngang",
  "chart.drawing.fibonacci": "Fibonacci",
  "chart.drawing.rectangle": "Hình chữ nhật",
  "chart.drawing.clear_all": "Xóa tất cả bản vẽ",

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
  "sector.VNCOM": "Truyền thông",

  // Dashboard
  "dashboard.title": "Bảng điều khiển",
  "dashboard.market_overview": "Tổng quan thị trường",
  "dashboard.hot_topics": "Chủ đề nóng",
  "dashboard.sector_performance": "Hiệu suất ngành",
  "dashboard.ma_crossover": "Phân bổ cắt MA",
  "dashboard.ai_valuation": "Xếp hạng định giá AI",
  "dashboard.global_markets": "Thị trường quốc tế",
  "dashboard.gold_price": "Giá vàng",
  "dashboard.gold_buy": "Mua",
  "dashboard.gold_sell": "Bán",
  "dashboard.fx_rates": "Tỷ giá ngoại tệ",
  "dashboard.interbank_rates": "Lãi suất liên ngân hàng",
  "dashboard.bond_yields": "Lợi suất trái phiếu",
  "dashboard.last_updated": "Cập nhật lần cuối: {{time}}",

  // Stock Analysis
  "stock.analysis": "Phân tích",
  "stock.overview": "Tổng quan",
  "stock.financials": "Tài chính",
  "stock.technical": "Kỹ thuật",
  "stock.news": "Tin tức",
  "stock.ai_thesis": "Luận điểm AI",
  "stock.valuation": "Định giá",
  "stock.officers": "Ban lãnh đạo",
  "stock.events": "Sự kiện",
  "stock.insider_trading": "Giao dịch nội bộ",
  "stock.order_book": "Sổ lệnh",
  "stock.bid": "Mua",
  "stock.ask": "Bán",
  "stock.match_price": "Giá khớp",
  "stock.match_volume": "KL khớp",
  "stock.support": "Hỗ trợ",
  "stock.resistance": "Kháng cự",
  "stock.signal": "Tín hiệu",
  "stock.strongly_bullish": "Rất tích cực",
  "stock.bullish": "Tích cực",
  "stock.neutral": "Trung lập",
  "stock.bearish": "Tiêu cực",
  "stock.strongly_bearish": "Rất tiêu cực",
  "stock.shareholders": "Cổ đông",
  "stock.subsidiaries": "Công ty con",
  "stock.company": "Doanh nghiệp",
  "stock.ownership": "Tỷ lệ sở hữu",
  "stock.shares": "Số cổ phần",
  "stock.shareholder_type": "Loại",
  "stock.industry": "Ngành",
  "stock.status": "Trạng thái",

  // Ranking
  "ranking.title": "Xếp hạng cổ phiếu AI",
  "ranking.factor_groups": "Nhóm yếu tố",
  "ranking.quality": "Chất lượng",
  "ranking.value": "Giá trị",
  "ranking.growth": "Tăng trưởng",
  "ranking.momentum": "Động lượng",
  "ranking.volatility_factor": "Biến động",
  "ranking.universe": "Phạm vi",
  "ranking.backtest": "Kiểm tra lại",
  "ranking.consensus_score": "Điểm đồng thuận",

  // Ideas
  "ideas.title": "Ý tưởng đầu tư",
  "ideas.buy_signal": "Tín hiệu mua",
  "ideas.sell_signal": "Tín hiệu bán",
  "ideas.confidence": "Độ tin cậy",
  "ideas.entry_price": "Giá vào",
  "ideas.target_price": "Giá mục tiêu",
  "ideas.stop_loss": "Cắt lỗ",
  "ideas.risk_reward": "Rủi ro/Lợi nhuận",

  // Missions
  "missions.title": "Nhiệm vụ",
  "missions.create": "Tạo nhiệm vụ",
  "missions.active": "Đang hoạt động",
  "missions.paused": "Tạm dừng",
  "missions.completed": "Hoàn thành",
  "missions.trigger": "Điều kiện kích hoạt",
  "missions.action": "Hành động",
  "missions.pause": "Tạm dừng",
  "missions.resume": "Tiếp tục",

  // Research
  "research.title": "Nghiên cứu",
  "research.factor_snapshot": "Ảnh chụp yếu tố",
  "research.sector_deepdive": "Phân tích ngành sâu",
  "research.market_outlook": "Triển vọng thị trường",
  "research.download_pdf": "Tải PDF",

  // Macro
  "macro.title": "Chỉ số vĩ mô",
  "macro.interbank_rate": "Lãi suất liên ngân hàng",
  "macro.bond_yield": "Lợi suất trái phiếu",
  "macro.fx_rate": "Tỷ giá",
  "macro.cpi": "CPI",
  "macro.gdp_growth": "Tăng trưởng GDP",
  "macro.exchange_rates": "Tỷ giá Vietcombank",
  "macro.gold": "Giá vàng",
  "macro.currency": "Tiền tệ",
  "macro.buy_rate": "Mua",
  "macro.transfer_rate": "Chuyển khoản",
  "macro.sell_rate": "Bán",

  // Heatmap
  "heatmap.title": "Bản đồ nhiệt thị trường",
  "heatmap.by_sector": "Theo ngành",
  "heatmap.by_market_cap": "Theo vốn hóa",

  // Chat / AI
  "chat.title": "Trợ lý AI",
  "chat.placeholder": "Hỏi về cổ phiếu, thị trường...",
  "chat.send": "Gửi",
  "chat.thinking": "Đang phân tích...",
  "chat.error": "Không thể xử lý yêu cầu. Vui lòng thử lại.",
  "chat.disclaimer": "Thông tin chỉ mang tính tham khảo, không phải lời khuyên đầu tư.",

  // Auth
  "auth.login": "Đăng nhập",
  "auth.logout": "Đăng xuất",
  "auth.username": "Tên đăng nhập",
  "auth.password": "Mật khẩu",
  "auth.login_failed": "Đăng nhập thất bại. Vui lòng kiểm tra lại thông tin.",
  "auth.logging_in": "Đang đăng nhập...",
  "auth.subtitle": "Nền tảng phân tích cổ phiếu Việt Nam",
  "auth.disclaimer_title": "Tuyên bố miễn trừ trách nhiệm",
  "auth.disclaimer_text": "Thông tin trên đây chỉ mang tính chất tham khảo, không phải lời khuyên đầu tư. EziStock không chịu trách nhiệm về quyết định đầu tư của bạn.",
  "auth.disclaimer_accept": "Tôi đã đọc và đồng ý",
  "auth.account_locked": "Tài khoản đã bị khóa. Vui lòng thử lại sau 30 phút.",
  "auth.register": "Đăng ký",
  "auth.register_failed": "Đăng ký thất bại. Tên đăng nhập có thể đã tồn tại.",
  "auth.no_account": "Chưa có tài khoản?",
  "auth.have_account": "Đã có tài khoản?",
  "auth.username_min": "Tên đăng nhập phải có ít nhất 3 ký tự",
  "auth.password_min": "Mật khẩu phải có ít nhất 8 ký tự",

  // Notifications
  "notifications.title": "Thông báo",
  "notifications.mark_all_read": "Đánh dấu tất cả đã đọc",
  "notifications.empty": "Không có thông báo mới",
  "notifications.price_alert": "Cảnh báo giá",
  "notifications.mission_triggered": "Nhiệm vụ kích hoạt",
  "notifications.idea_generated": "Ý tưởng mới",

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
  "btn.confirm": "Xác nhận",
  "btn.back": "Quay lại",
  "btn.next": "Tiếp theo",
  "btn.refresh": "Làm mới",

  // Errors
  "error.generic": "Đã xảy ra lỗi. Vui lòng thử lại.",
  "error.network": "Lỗi kết nối mạng.",
  "error.not_found": "Không tìm thấy dữ liệu.",
  "error.insufficient_holdings": "Số lượng nắm giữ không đủ.",
  "error.session_expired": "Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại.",
  "error.ai_unavailable": "Dịch vụ AI tạm thời không khả dụng.",

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
  "screener.load_preset": "Tải bộ lọc",
  "screener.liquidity_tier": "Hạng thanh khoản",

  // Settings
  "settings.title": "Cài đặt",
  "settings.language": "Ngôn ngữ",
  "settings.theme": "Giao diện",
  "settings.ai_config": "Cấu hình AI",
  "settings.notifications_config": "Cấu hình thông báo",
  "settings.theme_light": "Sáng",
  "settings.theme_dark": "Tối",

  // Help
  "help.tooltip_pe": "Tỷ lệ giá trên thu nhập",
  "help.tooltip_pb": "Tỷ lệ giá trên giá trị sổ sách",
  "help.tooltip_roe": "Tỷ suất lợi nhuận trên vốn chủ sở hữu",

  // Language selector
  "language.vi": "Tiếng Việt",
  "language.en": "English",

  // Data freshness
  "freshness.live": "Trực tiếp",
  "freshness.delayed": "Trễ",
  "freshness.stale": "Dữ liệu cũ",

  // Analyst
  "stock.analyst_consensus": "Đồng thuận phân tích",
  "stock.analysts": "nhà phân tích",
  "stock.buy_count": "Mua",
  "stock.hold_count": "Nắm giữ",
  "stock.sell_count": "Bán",
  "stock.accuracy": "Độ chính xác",

  // Chat suggestions
  "chat.suggestions": "Câu hỏi gợi ý",
  "chat.token_warning": "Sắp hết ngân sách token (80%)",

  // Terms
  "terms.title": "Điều khoản dịch vụ",
  "terms.disclaimer_title": "Tuyên bố miễn trừ trách nhiệm",
  "terms.disclaimer_text": "EziStock cung cấp phân tích cổ phiếu bằng AI chỉ nhằm mục đích tham khảo. Tất cả các khuyến nghị được tạo bởi hệ thống tự động và không nên được coi là lời khuyên tài chính. Luôn tham khảo ý kiến cố vấn tài chính được cấp phép trước khi đưa ra quyết định đầu tư.",
  "terms.risk_warning": "Bằng việc sử dụng EziStock, bạn thừa nhận rằng hiệu suất trong quá khứ không đảm bảo kết quả trong tương lai và tất cả các khoản đầu tư đều có rủi ro thua lỗ.",
  "terms.usage_title": "Điều khoản sử dụng",
  "terms.usage_1": "Bạn phải từ 18 tuổi trở lên để sử dụng nền tảng này.",
  "terms.usage_2": "Bạn chịu trách nhiệm về mọi quyết định đầu tư của mình.",
  "terms.usage_3": "Không được sử dụng nền tảng cho mục đích thao túng thị trường hoặc giao dịch nội gián.",
  "terms.privacy_title": "Chính sách bảo mật",
  "terms.privacy_text": "Chúng tôi cam kết bảo vệ quyền riêng tư của bạn. Dưới đây là cách chúng tôi xử lý dữ liệu của bạn:",
  "terms.privacy_1": "Khóa API và thông tin xác thực được lưu trữ cục bộ trên trình duyệt của bạn.",
  "terms.privacy_2": "Dữ liệu danh mục đầu tư được mã hóa khi lưu trữ và truyền tải.",
  "terms.privacy_3": "Chúng tôi không bán hoặc chia sẻ dữ liệu cá nhân của bạn cho bên thứ ba.",
  "terms.contact_title": "Liên hệ",
  "terms.contact_text": "Nếu bạn có câu hỏi về điều khoản dịch vụ, vui lòng liên hệ qua email: support@ezistock.vn",

  // Offline
  "offline.banner": "Chế độ ngoại tuyến",
  "offline.cached_data": "Hiển thị dữ liệu đã lưu",
  "offline.write_disabled": "Không thể thực hiện thao tác ghi khi ngoại tuyến",

  // Watchlist Manager
  "watchlist.new": "Danh sách mới",
  "watchlist.create": "Tạo",
  "watchlist.rename": "Đổi tên",
  "watchlist.delete_confirm": "Xóa \"{{name}}\"?",
  "watchlist.no_lists": "Chưa có danh sách theo dõi.",
  "watchlist.no_lists_hint": "Tạo một danh sách để bắt đầu theo dõi cổ phiếu.",
  "watchlist.no_symbols": "Chưa có mã nào trong danh sách.",
  "watchlist.add_symbol": "Thêm mã",
  "watchlist.alert_above": "Trên",
  "watchlist.alert_below": "Dưới",
  "watchlist.lists_count": "{{count}} danh sách",

  // Corporate Actions
  "corporate.title": "Sự kiện doanh nghiệp",
  "corporate.upcoming": "Sắp tới",
  "corporate.history": "Lịch sử",
  "corporate.no_upcoming": "Không có sự kiện sắp tới",
  "corporate.no_history": "Không có lịch sử cổ tức",
  "corporate.ex_date": "Ngày GDKHQ",
  "corporate.payment": "Ngày thanh toán",
  "corporate.amount": "Giá trị",
  "corporate.per_share": "/cp",
  "corporate.today": "Hôm nay",

  // Recommendation History
  "recommendations.title": "Lịch sử khuyến nghị",
  "recommendations.no_data": "Chưa có lịch sử khuyến nghị",
  "recommendations.outcome": "Kết quả",
  "recommendations.accuracy": "Độ chính xác",
  "recommendations.confidence": "Độ tin cậy",
  "recommendations.target": "Mục tiêu",
  "recommendations.actual": "Thực tế",
  "recommendations.hit": "Đúng",
  "recommendations.miss": "Sai",

  // Model Performance
  "model_perf.title": "Hiệu suất mô hình AI",
  "model_perf.agent_accuracy": "Xu hướng độ chính xác",
  "model_perf.hit_rate": "Tỷ lệ đúng theo độ tin cậy",
  "model_perf.predicted_vs_actual": "Dự đoán vs Thực tế",
  "model_perf.no_data": "Chưa có dữ liệu hiệu suất",
  "model_perf.regime": "Chế độ thị trường",
  "model_perf.accuracy_pct": "% Chính xác",

  // Market Status
  "market.status": "Trạng thái",
  "market.pre_open": "Trước giờ",
  "market.continuous": "Đang GD",
  "market.atc": "ATC",
  "market.break": "Nghỉ trưa",
  "market.closed": "Đóng cửa",

  // Common
  "common.loading": "Đang tải...",
  "common.no_data": "Không có dữ liệu",
  "common.search": "Tìm kiếm",
  "common.all": "Tất cả",
};

export default viVN;
