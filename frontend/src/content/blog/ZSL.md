---
title: "Zero-Shot Learning cho Phân Loại Ảnh Y Tế"
description: "Tài liệu yêu cầu hệ thống Zero-Shot Learning (ZSL) ứng dụng vào phân loại ảnh y tế - phân loại bệnh chưa từng thấy trong quá trình huấn luyện thông qua semantic knowledge transfer."
date: 2026-05-08
category: "nghien-cuu"
tags: ["zero-shot-learning", "machine-learning", "medical-imaging", "deep-learning", "computer-vision", "pytorch"]
password: "caothanhnguyen"
draft: false
coverImage: "https://images.unsplash.com/photo-1559757175-5700dde675bc?w=800&h=400&fit=crop"
---

# Tài Liệu Yêu Cầu (Requirements Document)

## Giới thiệu

Tài liệu này định nghĩa các yêu cầu cho hệ thống Zero-Shot Learning (ZSL) ứng dụng vào phân loại ảnh y tế, được phát triển trong khuôn khổ luận văn thạc sĩ. Hệ thống cho phép phân loại ảnh y tế vào các lớp bệnh chưa từng xuất hiện trong quá trình huấn luyện, thông qua cơ chế chuyển giao tri thức ngữ nghĩa (semantic knowledge transfer) sử dụng attribute embeddings và generative models. Dự án bao gồm toàn bộ pipeline nghiên cứu: tiền xử lý dữ liệu, thiết kế kiến trúc mô hình, huấn luyện trên seen classes, suy luận zero-shot trên unseen classes, và đánh giá toàn diện trong cả hai chế độ ZSL và Generalized ZSL (GZSL).

## Bảng thuật ngữ (Glossary)

- **ZSL_System**: Toàn bộ pipeline zero-shot learning cho phân loại ảnh y tế, bao gồm xử lý dữ liệu, trích xuất đặc trưng, semantic embedding, và các thành phần phân loại
- **Visual_Feature_Extractor**: Mạng neural tích chập sâu (deep CNN, ví dụ ResNet-101, DenseNet-121) trích xuất vector đặc trưng thị giác từ ảnh y tế
- **Semantic_Embedding_Module**: Thành phần ánh xạ nhãn lớp sang biểu diễn vector ngữ nghĩa sử dụng medical attributes hoặc text embeddings (ví dụ ClinicalBERT, BioBERT)
- **Compatibility_Function**: Hàm học được đo lường độ tương đồng giữa visual features và semantic embeddings để thực hiện phân loại
- **Generative_Module**: Mô hình sinh (ví dụ f-VAEGAN-D2, CADA-VAE) tổng hợp visual features cho unseen classes từ mô tả ngữ nghĩa của chúng
- **Data_Preprocessor**: Thành phần pipeline chịu trách nhiệm tải ảnh, augmentation, normalization, và chia lớp thành tập seen/unseen
- **Evaluation_Engine**: Thành phần tính toán các metrics phân loại bao gồm per-class accuracy, harmonic mean, và AUC scores
- **Seen_Classes**: Các lớp bệnh có sẵn trong quá trình huấn luyện với cả ảnh và mô tả ngữ nghĩa
- **Unseen_Classes**: Các lớp bệnh chỉ có sẵn tại thời điểm suy luận, được biểu diễn duy nhất bởi mô tả ngữ nghĩa
- **Attribute_Vector**: Vector số mã hóa các thuộc tính lâm sàng và thị giác của một lớp bệnh (ví dụ hình dạng tổn thương, màu sắc, vị trí, kết cấu)
- **GZSL**: Generalized Zero-Shot Learning — chế độ mà ảnh test có thể thuộc cả seen classes lẫn unseen classes
- **Harmonic_Mean**: Trung bình điều hòa của accuracy trên seen classes và unseen classes, được dùng làm metric chính cho GZSL

## Các yêu cầu (Requirements)

### Yêu cầu 1: Tải và tiền xử lý dữ liệu ảnh y tế

**User Story:** Với vai trò nghiên cứu sinh, tôi muốn tải và tiền xử lý các bộ dữ liệu ảnh y tế, để ảnh được chuẩn hóa và sẵn sàng cho bước trích xuất đặc trưng.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. WHEN đường dẫn dataset được cung cấp, THE Data_Preprocessor SHALL tải ảnh y tế từ các định dạng được hỗ trợ (DICOM, PNG, JPEG) và trả về normalized tensors
2. WHEN ảnh được tải, THE Data_Preprocessor SHALL resize tất cả ảnh về độ phân giải mục tiêu có thể cấu hình (mặc định 224x224 pixels)
3. WHEN dữ liệu huấn luyện được chuẩn bị, THE Data_Preprocessor SHALL áp dụng data augmentation bao gồm random horizontal flip, random rotation (tối đa 15 độ), và color jitter
4. THE Data_Preprocessor SHALL normalize giá trị pixel sử dụng mean và standard deviation của ImageNet
5. WHEN dataset được tải, THE Data_Preprocessor SHALL chia các lớp thành tập seen và unseen theo tỷ lệ có thể cấu hình (mặc định 70% seen, 30% unseen)
6. THE Data_Preprocessor SHALL đảm bảo không có ảnh nào từ unseen classes xuất hiện trong tập huấn luyện
7. IF gặp file ảnh bị hỏng hoặc không đọc được, THEN THE Data_Preprocessor SHALL ghi log cảnh báo và bỏ qua file đó mà không làm gián đoạn pipeline

### Yêu cầu 2: Xây dựng Semantic Embedding

**User Story:** Với vai trò nghiên cứu sinh, tôi muốn xây dựng semantic embeddings cho mỗi lớp bệnh, để mô hình có thể chuyển giao tri thức từ seen classes sang unseen classes.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Semantic_Embedding_Module SHALL hỗ trợ attribute-based embeddings trong đó mỗi lớp được biểu diễn bởi một Attribute_Vector các thuộc tính lâm sàng được định nghĩa thủ công
2. THE Semantic_Embedding_Module SHALL hỗ trợ text-based embeddings được sinh từ mô tả y tế của lớp bệnh sử dụng pre-trained language model (ClinicalBERT hoặc BioBERT)
3. WHEN một Attribute_Vector được xây dựng, THE Semantic_Embedding_Module SHALL mã hóa tối thiểu các thuộc tính sau: hình thái tổn thương (lesion morphology), mẫu màu sắc (color pattern), kết cấu (texture), vị trí giải phẫu (anatomical location), và tính đối xứng (symmetry)
4. THE Semantic_Embedding_Module SHALL tạo ra embedding vectors có chiều cố định cho tất cả các lớp (cả seen và unseen)
5. WHEN text-based embeddings được chọn, THE Semantic_Embedding_Module SHALL sinh embeddings từ mô tả y tế cấp lớp mà không yêu cầu annotation cho từng ảnh
6. IF một lớp không có mô tả ngữ nghĩa, THEN THE Semantic_Embedding_Module SHALL báo lỗi chỉ rõ lớp bị thiếu và dừng xử lý

### Yêu cầu 3: Trích xuất đặc trưng thị giác (Visual Feature Extraction)

**User Story:** Với vai trò nghiên cứu sinh, tôi muốn trích xuất các đặc trưng thị giác phân biệt từ ảnh y tế, để mô hình có thể học được mối liên hệ visual-semantic.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Visual_Feature_Extractor SHALL sử dụng pre-trained convolutional neural network (ResNet-101 hoặc DenseNet-121) làm backbone architecture
2. WHEN một ảnh được cung cấp, THE Visual_Feature_Extractor SHALL xuất ra feature vector có chiều cố định (2048 chiều cho ResNet-101)
3. THE Visual_Feature_Extractor SHALL hỗ trợ fine-tuning N layers cuối (có thể cấu hình) trong khi đóng băng (freeze) các layers trước đó
4. WHEN fine-tuning được bật, THE Visual_Feature_Extractor SHALL cập nhật trọng số backbone sử dụng training loss từ ZSL objective
5. THE Visual_Feature_Extractor SHALL trích xuất features từ penultimate layer (trước classification head) của backbone network
6. IF ảnh đầu vào có ít hơn 3 kênh màu, THEN THE Visual_Feature_Extractor SHALL nhân bản kênh để tạo đầu vào 3 kênh

### Yêu cầu 4: Mô hình sinh cho tổng hợp đặc trưng (Generative Model for Feature Synthesis)

**User Story:** Với vai trò nghiên cứu sinh, tôi muốn tổng hợp visual features cho unseen classes sử dụng generative models, để classifier có thể được huấn luyện trên cả features thật (seen) và features tổng hợp (unseen).

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Generative_Module SHALL triển khai ít nhất một kiến trúc sinh từ: f-VAEGAN-D2, CADA-VAE, hoặc TF-VAEGAN
2. WHEN được huấn luyện, THE Generative_Module SHALL nhận semantic embedding vector làm đầu vào và sinh ra synthetic visual feature vector
3. THE Generative_Module SHALL được huấn luyện sử dụng visual features của seen classes ghép cặp với semantic embeddings tương ứng
4. WHEN sinh features cho unseen classes, THE Generative_Module SHALL tạo ra synthetic features có phân phối thống kê nhất quán với phân phối visual features đã học
5. THE Generative_Module SHALL hỗ trợ sinh N mẫu tổng hợp cho mỗi unseen class (mặc định N=200, có thể cấu hình)
6. WHEN kiến trúc CADA-VAE được chọn, THE Generative_Module SHALL học một shared latent space căn chỉnh cả hai modalities (visual và semantic)
7. IF training loss của generative model không giảm trong 10 epochs liên tiếp, THEN THE Generative_Module SHALL kích hoạt early stopping và ghi log giá trị loss cuối cùng

### Yêu cầu 5: Phân loại Zero-Shot (Zero-Shot Classification)

**User Story:** Với vai trò nghiên cứu sinh, tôi muốn phân loại ảnh y tế vào các lớp bệnh chưa từng thấy, để hệ thống có thể chẩn đoán các tình trạng không có trong dữ liệu huấn luyện.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. WHEN hoạt động ở chế độ ZSL, THE Compatibility_Function SHALL phân loại ảnh test chỉ trong phạm vi unseen classes
2. WHEN hoạt động ở chế độ GZSL, THE Compatibility_Function SHALL phân loại ảnh test trong phạm vi cả seen và unseen classes
3. THE Compatibility_Function SHALL tính toán similarity scores giữa visual features và semantic embeddings của tất cả candidate classes
4. WHEN generative features có sẵn, THE ZSL_System SHALL huấn luyện softmax classifier trên features kết hợp (real từ seen + synthetic từ unseen)
5. THE Compatibility_Function SHALL hỗ trợ ít nhất hai phương pháp similarity: bilinear compatibility và cosine similarity trong learned projection space
6. WHEN calibration parameter được cấu hình, THE ZSL_System SHALL áp dụng calibrated stacking để giảm bias hướng về seen classes trong chế độ GZSL

### Yêu cầu 6: Chiến lược huấn luyện mô hình (Model Training Strategy)

**User Story:** Với vai trò nghiên cứu sinh, tôi muốn có pipeline huấn luyện có cấu trúc, để mô hình học được ánh xạ visual-semantic hiệu quả từ seen classes.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE ZSL_System SHALL huấn luyện chỉ sử dụng ảnh seen classes và semantic embeddings tương ứng trong giai đoạn huấn luyện chính
2. WHEN huấn luyện bắt đầu, THE ZSL_System SHALL khởi tạo Visual_Feature_Extractor với pre-trained ImageNet weights
3. THE ZSL_System SHALL hỗ trợ các hyperparameters có thể cấu hình bao gồm: learning rate, batch size, số epochs, weight decay, và latent dimension size
4. WHEN huấn luyện Generative_Module, THE ZSL_System SHALL tối ưu hóa combined loss gồm reconstruction loss, KL-divergence, và adversarial loss (cho các biến thể VAEGAN)
5. THE ZSL_System SHALL triển khai learning rate scheduling với configurable step size và decay factor
6. THE ZSL_System SHALL lưu model checkpoints theo khoảng thời gian có thể cấu hình và giữ lại checkpoint tốt nhất dựa trên validation accuracy
7. IF GPU memory không đủ cho batch size đã cấu hình, THEN THE ZSL_System SHALL tự động giảm batch size và ghi log điều chỉnh

### Yêu cầu 7: Đánh giá và Metrics (Evaluation and Metrics)

**User Story:** Với vai trò nghiên cứu sinh, tôi muốn có các metrics đánh giá toàn diện, để tôi có thể đánh giá hiệu suất mô hình và so sánh với các phương pháp baseline.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE Evaluation_Engine SHALL tính toán per-class top-1 accuracy cho tất cả unseen classes trong chế độ ZSL
2. THE Evaluation_Engine SHALL tính toán Harmonic_Mean của seen-class accuracy (S) và unseen-class accuracy (U) cho chế độ GZSL theo công thức: H = 2*S*U / (S+U)
3. THE Evaluation_Engine SHALL tính toán Area Under the ROC Curve (AUC) cho các kịch bản phân loại binary và multi-class
4. WHEN đánh giá hoàn tất, THE Evaluation_Engine SHALL sinh classification report bao gồm per-class precision, recall, và F1-score
5. THE Evaluation_Engine SHALL hỗ trợ so sánh với ít nhất ba phương pháp baseline: random chance, nearest-neighbor trong semantic space, và direct attribute prediction baseline
6. WHEN đánh giá được chạy, THE Evaluation_Engine SHALL tạo confusion matrices cho cả seen và unseen class predictions
7. THE Evaluation_Engine SHALL ghi tất cả metrics vào file đầu ra có cấu trúc (định dạng JSON) để đảm bảo tính tái tạo

### Yêu cầu 8: Cấu hình thí nghiệm và tính tái tạo (Experiment Configuration and Reproducibility)

**User Story:** Với vai trò nghiên cứu sinh, tôi muốn các thí nghiệm có thể tái tạo với tham số có thể cấu hình, để tôi có thể so sánh hệ thống các phương pháp khác nhau trong luận văn.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. THE ZSL_System SHALL tải tất cả tham số thí nghiệm từ file cấu hình YAML
2. THE ZSL_System SHALL đặt random seeds cho Python, NumPy, và PyTorch để đảm bảo kết quả có thể tái tạo
3. WHEN một thí nghiệm bắt đầu, THE ZSL_System SHALL ghi log toàn bộ cấu hình, git commit hash (nếu có), và chi tiết môi trường
4. THE ZSL_System SHALL hỗ trợ nhiều cấu hình dataset bao gồm CheXpert (X-quang ngực), ISIC (dermoscopy), và các biến thể MedMNIST (PathMNIST, DermaMNIST)
5. WHEN một thí nghiệm mới được khởi chạy, THE ZSL_System SHALL tạo thư mục đầu ra có timestamp chứa tất cả logs, checkpoints, và kết quả đánh giá
6. THE ZSL_System SHALL hỗ trợ command-line overrides cho bất kỳ tham số cấu hình nào mà không cần sửa file cấu hình

### Yêu cầu 9: Trực quan hóa và phân tích (Visualization and Analysis)

**User Story:** Với vai trò nghiên cứu sinh, tôi muốn có công cụ trực quan hóa để phân tích, để tôi có thể diễn giải hành vi mô hình và đưa kết quả vào luận văn.

#### Tiêu chí chấp nhận (Acceptance Criteria)

1. WHEN được yêu cầu, THE ZSL_System SHALL sinh trực quan hóa t-SNE hoặc UMAP của embedding space đã học, hiển thị cả seen và unseen class clusters
2. THE ZSL_System SHALL tạo training curves vẽ loss và accuracy theo epochs
3. WHEN đánh giá hoàn tất, THE ZSL_System SHALL sinh bar charts so sánh per-class accuracy giữa seen và unseen classes
4. THE ZSL_System SHALL hỗ trợ Grad-CAM visualization để highlight các vùng ảnh đóng góp vào quyết định phân loại
5. WHEN Generative_Module được huấn luyện, THE ZSL_System SHALL trực quan hóa phân phối của real features so với synthetic features sử dụng dimensionality reduction
6. THE ZSL_System SHALL xuất tất cả hình ảnh ở định dạng chất lượng xuất bản (PDF và PNG tối thiểu 300 DPI)
