# Requirements Document

## Introduction

Tính năng nội bộ "Lịch ăn & Chia thanh toán" dành cho nhóm làm blog. Nhóm có nhiều thành viên cùng ăn sáng/trưa, mỗi bữa sẽ có 1 người đứng ra thanh toán. Hệ thống tự động phân chia lịch thanh toán sao cho công bằng nhất — người nào đã thanh toán nhiều sẽ được ưu tiên nghỉ, người chưa thanh toán hoặc thanh toán ít sẽ được xếp lịch tiếp theo.

Tính năng này là internal tool, chỉ nhóm blog truy cập được (không public). Dữ liệu lưu trên backend (PostgreSQL), giao diện tích hợp vào frontend blog hiện tại.

## Requirements

### Requirement 1: Quản lý thành viên nhóm

**User Story:** As a nhóm trưởng, I want to quản lý danh sách thành viên trong nhóm, so that hệ thống biết ai tham gia ăn và ai cần được xếp lịch thanh toán.

#### Acceptance Criteria

1. WHEN người dùng truy cập trang quản lý nhóm THEN hệ thống SHALL hiển thị danh sách tất cả thành viên với tên và trạng thái (active/inactive).
2. WHEN người dùng thêm thành viên mới với tên hợp lệ THEN hệ thống SHALL lưu thành viên vào database với trạng thái active.
3. WHEN người dùng đặt thành viên thành inactive THEN hệ thống SHALL loại thành viên đó khỏi lịch xếp thanh toán nhưng vẫn giữ lịch sử.
4. IF tên thành viên trùng với thành viên đã tồn tại THEN hệ thống SHALL từ chối và hiển thị thông báo lỗi.
5. WHEN người dùng xóa thành viên THEN hệ thống SHALL xóa mềm (soft delete) và giữ lại toàn bộ lịch sử thanh toán.

### Requirement 2: Cấu hình bữa ăn

**User Story:** As a thành viên nhóm, I want to cấu hình các bữa ăn (sáng, trưa) và ai tham gia bữa nào, so that lịch thanh toán chỉ áp dụng cho những người thực sự ăn.

#### Acceptance Criteria

1. WHEN hệ thống khởi tạo THEN hệ thống SHALL hỗ trợ 2 loại bữa ăn: sáng (breakfast) và trưa (lunch).
2. WHEN thành viên đăng ký tham gia bữa ăn THEN hệ thống SHALL ghi nhận thành viên đó vào danh sách tham gia bữa tương ứng.
3. WHEN thành viên hủy tham gia bữa ăn THEN hệ thống SHALL loại thành viên khỏi lịch xếp thanh toán của bữa đó.
4. IF thành viên không tham gia bữa nào THEN hệ thống SHALL không xếp lịch thanh toán cho thành viên đó.
5. WHEN người dùng xem cấu hình bữa ăn THEN hệ thống SHALL hiển thị danh sách thành viên tham gia mỗi bữa.

### Requirement 3: Thuật toán xếp lịch thanh toán công bằng

**User Story:** As a thành viên nhóm, I want to hệ thống tự động xếp lịch ai thanh toán bữa tiếp theo, so that mọi người đều được chia đều số lần thanh toán.

#### Acceptance Criteria

1. WHEN hệ thống xếp lịch thanh toán cho bữa tiếp theo THEN hệ thống SHALL chọn thành viên có số lần thanh toán ít nhất trong bữa đó.
2. IF có nhiều thành viên cùng số lần thanh toán ít nhất THEN hệ thống SHALL chọn thành viên có lần thanh toán gần nhất xa nhất (lâu nhất chưa trả).
3. IF có nhiều thành viên cùng số lần thanh toán và cùng thời gian THEN hệ thống SHALL chọn ngẫu nhiên trong số đó.
4. WHEN thành viên mới tham gia THEN hệ thống SHALL bắt đầu đếm từ 0 lần thanh toán cho thành viên đó.
5. WHEN hệ thống hiển thị lịch thanh toán THEN hệ thống SHALL hiển thị người thanh toán tiếp theo cho mỗi bữa (sáng, trưa).
6. WHEN người dùng xem thống kê THEN hệ thống SHALL hiển thị số lần thanh toán của mỗi thành viên theo từng bữa.

### Requirement 4: Ghi nhận thanh toán

**User Story:** As a thành viên nhóm, I want to ghi nhận khi ai đó đã thanh toán bữa ăn, so that hệ thống cập nhật lịch sử và tính toán lịch tiếp theo chính xác.

#### Acceptance Criteria

1. WHEN người dùng xác nhận thanh toán THEN hệ thống SHALL ghi nhận thành viên đã thanh toán, bữa ăn, ngày, và cập nhật số lần thanh toán.
2. WHEN thanh toán được ghi nhận THEN hệ thống SHALL tự động tính toán lại người thanh toán tiếp theo.
3. IF người thanh toán thực tế khác với người được xếp lịch THEN hệ thống SHALL cho phép chọn người thanh toán thực tế (override).
4. WHEN người dùng xem lịch sử thanh toán THEN hệ thống SHALL hiển thị danh sách các lần thanh toán theo thứ tự thời gian (mới nhất trước).
5. WHEN người dùng muốn hủy ghi nhận thanh toán gần nhất THEN hệ thống SHALL cho phép undo lần ghi nhận cuối cùng trong vòng 24 giờ.

### Requirement 5: Dashboard hiển thị

**User Story:** As a thành viên nhóm, I want to xem dashboard tổng quan lịch ăn và thanh toán, so that tôi biết ai thanh toán tiếp và tình hình chia đều hiện tại.

#### Acceptance Criteria

1. WHEN người dùng truy cập dashboard THEN hệ thống SHALL hiển thị người thanh toán tiếp theo cho bữa sáng và bữa trưa.
2. WHEN người dùng truy cập dashboard THEN hệ thống SHALL hiển thị bảng thống kê số lần thanh toán của mỗi thành viên theo từng bữa.
3. WHEN người dùng truy cập dashboard THEN hệ thống SHALL hiển thị lịch sử 10 lần thanh toán gần nhất.
4. WHEN dữ liệu thay đổi (thêm thanh toán, thêm thành viên) THEN hệ thống SHALL cập nhật dashboard realtime (hoặc sau khi refresh).
5. WHEN người dùng truy cập trên mobile THEN hệ thống SHALL hiển thị dashboard responsive, dễ đọc trên màn hình nhỏ.

### Requirement 6: Bảo mật nội bộ

**User Story:** As a nhóm trưởng, I want to tính năng này chỉ nhóm nội bộ truy cập được, so that người ngoài không thể xem hoặc thay đổi dữ liệu nhóm.

#### Acceptance Criteria

1. WHEN người dùng truy cập trang meal scheduler THEN hệ thống SHALL yêu cầu nhập mật khẩu nhóm (shared password).
2. WHEN mật khẩu đúng THEN hệ thống SHALL lưu trạng thái đăng nhập vào sessionStorage và cho phép truy cập.
3. WHEN mật khẩu sai THEN hệ thống SHALL hiển thị thông báo lỗi và không cho truy cập.
4. IF sessionStorage hết hạn hoặc bị xóa THEN hệ thống SHALL yêu cầu nhập lại mật khẩu.
5. WHEN API endpoint được gọi THEN hệ thống SHALL xác thực request bằng shared secret trong header (X-Group-Secret).
6. IF request không có hoặc sai secret THEN hệ thống SHALL trả về 401 Unauthorized.
