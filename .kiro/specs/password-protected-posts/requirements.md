# Tài liệu Yêu cầu

## Giới thiệu

Tính năng bảo vệ bài viết bằng mật khẩu cho phép chủ blog đặt mật khẩu cho một số bài viết nhất định. Người đọc cần nhập đúng mật khẩu mới có thể xem nội dung bài viết. Giải pháp hoàn toàn xử lý ở frontend (client-side), không yêu cầu thay đổi backend. Mật khẩu được cấu hình trong frontmatter của bài viết Markdown. Blog sử dụng Astro (static output) với Svelte components và TailwindCSS.

## Thuật ngữ

- **Password_Gate**: Svelte component hiển thị form nhập mật khẩu, chặn người đọc truy cập nội dung bài viết cho đến khi nhập đúng mật khẩu
- **Blog_Post_Page**: Trang chi tiết bài viết (`[...slug].astro`), nơi hiển thị nội dung Markdown đã render
- **Content_Schema**: Zod schema định nghĩa cấu trúc frontmatter của blog collection trong `content/config.ts`
- **Frontmatter**: Phần metadata YAML ở đầu file Markdown, chứa các trường như title, date, password
- **Session_Storage**: Web Storage API lưu trữ dữ liệu trong phiên trình duyệt hiện tại, tự động xóa khi đóng tab/trình duyệt
- **Blog_Listing**: Các trang hiển thị danh sách bài viết (trang chủ, trang category, trang tag, trang tìm kiếm)
- **Protected_Post**: Bài viết có trường `password` được cấu hình trong frontmatter

## Yêu cầu

### Yêu cầu 1: Cấu hình mật khẩu trong Frontmatter

**Câu chuyện người dùng:** Là chủ blog, tôi muốn đặt mật khẩu cho các bài viết cụ thể trong frontmatter, để tôi có thể kiểm soát bài viết nào yêu cầu xác thực mới được đọc.

#### Tiêu chí chấp nhận

1. Content_Schema PHẢI bao gồm trường `password` tùy chọn (optional) kiểu string với độ dài tối đa 128 ký tự trong blog collection schema
2. KHI một bài viết có trường `password` được định nghĩa với giá trị chuỗi không rỗng và không chỉ chứa khoảng trắng trong Frontmatter, Blog_Post_Page PHẢI coi bài viết đó là Protected_Post bằng cách ẩn nội dung và hiển thị form nhập mật khẩu thay thế
3. KHI một bài viết không có trường `password` trong Frontmatter, Blog_Post_Page PHẢI hiển thị nội dung bài viết mà không có bất kỳ password gate nào
4. NẾU trường `password` trong Frontmatter được đặt thành chuỗi rỗng, undefined, hoặc chuỗi chỉ chứa khoảng trắng, THÌ Content_Schema PHẢI coi bài viết là không được bảo vệ và Blog_Post_Page PHẢI hiển thị nội dung bài viết mà không có password gate
5. KHI một bài viết là Protected_Post, Blog_Post_Page PHẢI hiển thị tiêu đề và metadata của bài viết nhưng KHÔNG ĐƯỢC render nội dung bài viết cho đến khi mật khẩu đúng được cung cấp

### Yêu cầu 2: Hiển thị Password Gate

**Câu chuyện người dùng:** Là người đọc blog, tôi muốn thấy form nhập mật khẩu rõ ràng khi truy cập bài viết được bảo vệ, để tôi biết bài viết yêu cầu mật khẩu và có thể nhập nó.

#### Tiêu chí chấp nhận

1. KHI người đọc truy cập vào một Protected_Post, Password_Gate PHẢI hiển thị form nhập mật khẩu thay vì nội dung bài viết
2. Password_Gate PHẢI hiển thị biểu tượng khóa và thông báo cho biết bài viết được bảo vệ bằng mật khẩu
3. Password_Gate PHẢI bao gồm trường nhập liệu kiểu "password" để nhập mật khẩu, với độ dài nhập tối đa 128 ký tự
4. Password_Gate PHẢI bao gồm nút gửi (submit) để xác minh mật khẩu đã nhập
5. Password_Gate PHẢI hỗ trợ gửi form bằng phím Enter khi đang ở trường nhập mật khẩu
6. Password_Gate PHẢI hiển thị tiêu đề bài viết và metadata có sẵn (ngày, danh mục, thẻ) phía trên form mật khẩu để người đọc xác nhận đúng bài viết; NẾU một trường metadata (danh mục hoặc thẻ) không được định nghĩa cho bài viết, THÌ Password_Gate PHẢI bỏ qua trường đó mà không hiển thị placeholder trống
7. KHI Password_Gate đang kiểm tra Session_Storage để tìm trạng thái đã mở khóa trước đó, Password_Gate KHÔNG ĐƯỢC hiển thị chớp nhoáng form mật khẩu trước khi hiện nội dung

### Yêu cầu 3: Xác thực mật khẩu phía Client

**Câu chuyện người dùng:** Là người đọc blog, tôi muốn mở khóa bài viết được bảo vệ bằng cách nhập đúng mật khẩu, để tôi có thể đọc nội dung.

#### Tiêu chí chấp nhận

1. KHI người đọc gửi đúng mật khẩu, Password_Gate PHẢI ẩn form mật khẩu và render toàn bộ nội dung bài viết thay thế
2. KHI người đọc gửi sai mật khẩu, Password_Gate PHẢI hiển thị thông báo lỗi cho biết mật khẩu không đúng, và thông báo lỗi PHẢI hiển thị cho đến khi người đọc gửi lại
3. KHI người đọc gửi sai mật khẩu, Password_Gate PHẢI xóa trường nhập mật khẩu và giữ focus trên trường nhập
4. NẾU trường nhập mật khẩu trống hoặc chỉ chứa ký tự khoảng trắng, THÌ Password_Gate PHẢI vô hiệu hóa nút gửi
5. Password_Gate PHẢI so sánh mật khẩu đã nhập với giá trị mật khẩu lưu trong Frontmatter của bài viết bằng phép so sánh chuỗi chính xác (strict equality) mà không cắt khoảng trắng ở hai đầu
6. Password_Gate PHẢI chấp nhận mật khẩu nhập vào với độ dài tối đa 128 ký tự

### Yêu cầu 4: Lưu trạng thái đã mở khóa trong Session

**Câu chuyện người dùng:** Là người đọc blog, tôi muốn giữ trạng thái đã mở khóa trên bài viết được bảo vệ trong phiên duyệt web, để tôi không phải nhập lại mật khẩu mỗi khi quay lại cùng bài viết.

#### Tiêu chí chấp nhận

1. KHI người đọc mở khóa thành công một Protected_Post, Password_Gate PHẢI lưu trạng thái đã mở khóa vào Session_Storage với định dạng key `protected-post:{slug}` trong đó `{slug}` là slug của bài viết
2. KHI người đọc truy cập vào một Protected_Post đã được mở khóa trước đó trong phiên hiện tại, Password_Gate PHẢI hiển thị nội dung bài viết mà không yêu cầu nhập lại mật khẩu
3. KHI tab hoặc cửa sổ trình duyệt được đóng, Session_Storage PHẢI tự động xóa tất cả trạng thái đã mở khóa
4. Password_Gate PHẢI chỉ lưu trạng thái mở khóa (giá trị boolean `true`) trong Session_Storage, không lưu giá trị mật khẩu
5. NẾU Session_Storage không khả dụng (ví dụ: chế độ duyệt riêng tư với storage bị tắt), THÌ Password_Gate PHẢI yêu cầu nhập mật khẩu mỗi lần tải trang mà không hiển thị lỗi

### Yêu cầu 5: Hiển thị bài viết được bảo vệ trong danh sách

**Câu chuyện người dùng:** Là người đọc blog, tôi muốn thấy các bài viết được bảo vệ trong danh sách blog với chỉ báo trực quan, để tôi biết bài viết nào yêu cầu mật khẩu trước khi nhấp vào.

#### Tiêu chí chấp nhận

1. Blog_Listing PHẢI hiển thị các mục Protected_Post cùng với các bài viết thường, sắp xếp theo ngày giảm dần (mới nhất trước), sử dụng cùng logic sắp xếp như bài viết thường
2. Blog_Listing PHẢI hiển thị biểu tượng khóa bên cạnh tiêu đề của mỗi mục Protected_Post để phân biệt với bài viết thường
3. Blog_Listing PHẢI hiển thị tiêu đề và mô tả của các mục Protected_Post giống hệt bài viết thường, mà không để lộ nội dung bài viết
4. KHI người đọc nhấp vào một mục Protected_Post trong Blog_Listing, Blog_Listing PHẢI điều hướng người đọc đến Blog_Post_Page nơi Password_Gate được hiển thị
5. NẾU một Protected_Post đã được mở khóa trước đó trong phiên hiện tại qua Session_Storage, THÌ Blog_Listing VẪN PHẢI hiển thị biểu tượng khóa trên mục bài viết đó
6. Blog_Listing PHẢI hiển thị biểu tượng khóa trên các mục Protected_Post trong tất cả ngữ cảnh danh sách bao gồm trang chủ, trang danh mục, trang thẻ, và kết quả tìm kiếm

### Yêu cầu 6: Khả năng truy cập và Trải nghiệm người dùng

**Câu chuyện người dùng:** Là người đọc blog sử dụng công nghệ hỗ trợ, tôi muốn password gate hoàn toàn có thể truy cập được, để tôi có thể tương tác với nó bất kể khả năng của mình.

#### Tiêu chí chấp nhận

1. Password_Gate PHẢI bao gồm thuộc tính `aria-label` với giá trị "Nhập mật khẩu để mở khóa bài viết này" trên trường nhập mật khẩu, và thuộc tính `aria-label` với giá trị "Mở khóa bài viết" trên nút gửi
2. Password_Gate PHẢI thông báo thông điệp lỗi cho trình đọc màn hình bằng container có thuộc tính `role="alert"` và `aria-live="assertive"`
3. Password_Gate PHẢI hoàn toàn điều hướng được bằng bàn phím: phím Tab để di chuyển giữa input và button, phím Enter để gửi form, và phím Escape để xóa trường nhập
4. Password_Gate PHẢI sử dụng màu chữ duy trì tỷ lệ tương phản tối thiểu 4.5:1 so với nền trong cả theme sáng và tối, tuân thủ WCAG 2.1 Level AA
5. KHI Password_Gate được hiển thị, Password_Gate PHẢI tự động focus vào trường nhập mật khẩu để tương tác ngay lập tức

### Yêu cầu 7: Bảo mật cơ bản phía Client

**Câu chuyện người dùng:** Là chủ blog, tôi muốn bảo vệ cơ bản phía client cho bài viết, để người đọc thông thường không thể truy cập nội dung mà không có mật khẩu.

#### Tiêu chí chấp nhận

1. TRONG KHI Password_Gate đang hiển thị, Blog_Post_Page KHÔNG ĐƯỢC bao gồm nội dung bài viết trong cây DOM đã render
2. Blog_Post_Page PHẢI render nội dung bài viết chỉ sau khi xác minh mật khẩu thành công ở phía client, bằng cách render có điều kiện khối nội dung trong Svelte component
3. Password_Gate PHẢI truyền mật khẩu vào Svelte component dưới dạng prop mà không thể truy cập trực tiếp qua `document.querySelector` hoặc biến JavaScript toàn cục trong console trình duyệt
4. NẾU người đọc cố gắng bỏ qua Password_Gate bằng cách tắt JavaScript, THÌ Blog_Post_Page KHÔNG ĐƯỢC hiển thị nội dung bài viết vì Svelte component yêu cầu JavaScript để render
