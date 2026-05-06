---
title: "Kiến Thức Nền Tảng Trước Khi Chọn Cloud"
description: "Học Linux, Docker, SSH và các kiến thức cơ bản trước khi chọn AWS, Oracle Cloud hay GCP. Giống như học C trước Java - nền tảng trước, công cụ sau!"
date: 2026-05-06
category: "cong-nghe"
tags: ["linux", "docker", "ssh", "devops", "cloud", "aws", "oracle-cloud", "tutorial"]
draft: false
coverImage: "https://images.unsplash.com/photo-1451187580459-43490279c0fa?w=800&h=400&fit=crop"
---

> *"Giống như học C trước Java - Học nền tảng trước, công cụ sau!"*

## 🎯 Triết Lý

```
C → Java
Nền tảng → Công cụ
Base Knowledge → Cloud Platform

Tại sao?
→ C dạy bạn cách máy tính hoạt động
→ Java dạy bạn cách xây dựng ứng dụng

Tương tự:
→ Linux/Docker dạy bạn cách server hoạt động
→ AWS/Oracle dạy bạn cách dùng cloud platform
```

---

## 📊 LỘ TRÌNH HỌC

### ❌ Cách Học SAI (Phổ biến):

```
Ngày 1: Đăng ký AWS
Ngày 2: Click button trên AWS Console
Ngày 3: Deploy app (không hiểu gì)
Ngày 4: Chuyển sang Oracle → Bối rối hoàn toàn

Vấn đề:
→ Không hiểu bản chất
→ Phụ thuộc vào UI
→ Chuyển cloud = học lại từ đầu
```

---

### ✅ Cách Học ĐÚNG (Khuyến nghị):

```
GIAI ĐOẠN 1: Kiến Thức Nền Tảng (2-3 tháng)
├─ Linux Commands
├─ SSH & Remote Access
├─ Docker & Containerization
├─ Git & Version Control
├─ Networking Basics
└─ Troubleshooting

GIAI ĐOẠN 2: Chọn Cloud (1 tháng)
├─ Hiểu đã vững base
├─ Chọn cloud phù hợp
└─ Apply kiến thức base vào cloud

Kết quả:
→ Hiểu bản chất
→ Độc lập với UI
→ Chuyển cloud dễ dàng (15 phút)
```

---

## 📚 PHẦN 1: KIẾN THỨC NỀN TẢNG

> **Mục tiêu:** Hiểu bản chất, không phụ thuộc cloud nào

---

## 🐧 CHƯƠNG 1: LINUX COMMANDS

### Tại Sao Học Linux?

```
✅ 90% server chạy Linux
✅ Mọi cloud đều dùng Linux
✅ Không biết Linux = không làm được DevOps
✅ Interview luôn hỏi Linux
```

---

### 1.1: Điều Hướng (Navigation)

**Mục tiêu:** Di chuyển giữa các thư mục

```bash
# Xem thư mục hiện tại
pwd
# Output: /home/ubuntu

# Liệt kê files
ls
ls -la  # Chi tiết hơn

# Di chuyển vào thư mục
cd /home/ubuntu/my-app
cd ..  # Lên 1 cấp
cd ~   # Về home directory

# Tạo thư mục
mkdir my-folder
mkdir -p parent/child/grandchild  # Tạo nhiều cấp

# Xóa thư mục
rmdir my-folder  # Chỉ xóa thư mục rỗng
rm -rf my-folder  # Xóa thư mục và nội dung (NGUY HIỂM!)
```

**Bài Tập:**

```bash
# 1. Tạo cấu trúc thư mục
mkdir -p projects/my-app/src
mkdir -p projects/my-app/docs

# 2. Di chuyển vào src
cd projects/my-app/src

# 3. Kiểm tra đang ở đâu
pwd

# 4. Về home
cd ~

# 5. Xóa projects
rm -rf projects
```

---

### 1.2: File Operations

**Mục tiêu:** Tạo, xem, sửa, xóa files

```bash
# Tạo file rỗng
touch file.txt

# Tạo file với nội dung
echo "Hello World" > file.txt

# Thêm nội dung vào file
echo "Line 2" >> file.txt

# Xem nội dung file
cat file.txt
less file.txt  # Xem từng trang
head file.txt  # 10 dòng đầu
tail file.txt  # 10 dòng cuối
tail -f file.txt  # Xem real-time (logs)

# Copy file
cp file.txt file-backup.txt
cp -r folder/ folder-backup/  # Copy thư mục

# Di chuyển/đổi tên file
mv file.txt newname.txt
mv file.txt /path/to/destination/

# Xóa file
rm file.txt
rm -f file.txt  # Force delete
```

**Bài Tập:**

```bash
# 1. Tạo file
echo "My first file" > test.txt

# 2. Xem nội dung
cat test.txt

# 3. Thêm dòng mới
echo "Second line" >> test.txt

# 4. Copy file
cp test.txt test-backup.txt

# 5. Đổi tên
mv test.txt renamed.txt

# 6. Xóa
rm renamed.txt test-backup.txt
```

---

### 1.3: Text Editors

**Mục tiêu:** Sửa file trên server

#### Nano (Dễ nhất - Khuyến nghị cho beginners)

```bash
# Mở file
nano file.txt

# Shortcuts:
Ctrl + O  # Save (WriteOut)
Enter     # Confirm
Ctrl + X  # Exit
Ctrl + K  # Cut line
Ctrl + U  # Paste
Ctrl + W  # Search
```

#### Vim (Mạnh hơn - Cho advanced)

```bash
# Mở file
vim file.txt

# Modes:
i         # Insert mode (để gõ)
Esc       # Normal mode
:w        # Save
:q        # Quit
:wq       # Save and quit
:q!       # Quit without save

# Navigation (Normal mode):
h j k l   # Left, Down, Up, Right
dd        # Delete line
yy        # Copy line
p         # Paste
```

**Bài Tập:**

```bash
# 1. Tạo file với nano
nano myfile.txt
# Gõ: "Hello from nano"
# Ctrl+O, Enter, Ctrl+X

# 2. Xem file
cat myfile.txt

# 3. Sửa file
nano myfile.txt
# Thêm dòng: "Second line"
# Ctrl+O, Enter, Ctrl+X
```

---

### 1.4: Permissions

**Mục tiêu:** Quản lý quyền truy cập files

```bash
# Xem permissions
ls -la
# Output: -rw-r--r-- 1 ubuntu ubuntu 1234 May 01 10:00 file.txt
#         ↑         ↑      ↑      ↑    ↑
#         type      owner  group  size date

# Permission format:
# -rw-r--r--
#  ↑↑↑ ↑↑↑ ↑↑↑
#  │   │   └─ Others (read, read, read)
#  │   └───── Group (read, write, execute)
#  └───────── Owner (read, write, execute)

# Đổi permissions
chmod 644 file.txt  # rw-r--r--
chmod 755 file.txt  # rwxr-xr-x
chmod 600 file.txt  # rw-------
chmod +x script.sh  # Thêm execute permission

# Đổi owner
chown ubuntu:ubuntu file.txt
chown -R ubuntu:ubuntu folder/  # Recursive
```

**Số Permission:**

```
r (read)    = 4
w (write)   = 2
x (execute) = 1

644 = rw-r--r--
    = 6 (4+2) owner: read+write
    = 4 (4)   group: read
    = 4 (4)   others: read

755 = rwxr-xr-x
    = 7 (4+2+1) owner: read+write+execute
    = 5 (4+1)   group: read+execute
    = 5 (4+1)   others: read+execute
```

**Bài Tập:**

```bash
# 1. Tạo file
touch secret.txt

# 2. Xem permissions
ls -la secret.txt

# 3. Chỉ owner đọc/ghi
chmod 600 secret.txt

# 4. Kiểm tra
ls -la secret.txt
# Output: -rw------- ...

# 5. Tạo script
echo '#!/bin/bash' > script.sh
echo 'echo "Hello"' >> script.sh

# 6. Thêm execute permission
chmod +x script.sh

# 7. Chạy script
./script.sh
```

---

### 1.5: Package Management

**Mục tiêu:** Cài đặt software

```bash
# Ubuntu/Debian (apt)
sudo apt update              # Update package list
sudo apt upgrade -y          # Upgrade packages
sudo apt install nginx -y    # Install package
sudo apt remove nginx        # Remove package
sudo apt autoremove          # Remove unused packages

# CentOS/RedHat (yum)
sudo yum update
sudo yum install nginx -y
sudo yum remove nginx
```

**Bài Tập:**

```bash
# 1. Update system
sudo apt update

# 2. Cài curl
sudo apt install curl -y

# 3. Test curl
curl --version

# 4. Cài htop (process monitor)
sudo apt install htop -y

# 5. Chạy htop
htop
# Press q to quit
```

---

### 1.6: Process Management

**Mục tiêu:** Quản lý processes đang chạy

```bash
# Xem processes
ps aux
ps aux | grep nginx  # Tìm process cụ thể

# Xem processes real-time
top
htop  # Đẹp hơn (cần cài: sudo apt install htop)

# Kill process
kill PID
kill -9 PID  # Force kill

# Tìm PID
pgrep nginx
pidof nginx

# Chạy process background
command &
nohup command &  # Chạy ngay cả khi logout
```

**Bài Tập:**

```bash
# 1. Xem tất cả processes
ps aux

# 2. Chạy process background
sleep 100 &

# 3. Xem process vừa tạo
ps aux | grep sleep

# 4. Kill process
kill $(pgrep sleep)

# 5. Kiểm tra đã kill
ps aux | grep sleep
```

---

### 1.7: Network Commands

**Mục tiêu:** Kiểm tra network

```bash
# Test connectivity
ping google.com
ping -c 4 google.com  # Ping 4 lần

# Download file
wget https://example.com/file.zip
curl -O https://example.com/file.zip

# Test API
curl http://localhost:8080/api/profiles
curl -X POST http://localhost:8080/api/profiles \
  -H "Content-Type: application/json" \
  -d '{"name":"Test"}'

# Xem ports đang mở
netstat -tulpn
ss -tulpn  # Modern alternative

# Xem IP address
ip addr
ifconfig  # Old command
```

**Bài Tập:**

```bash
# 1. Ping Google
ping -c 4 google.com

# 2. Download file
curl -O https://raw.githubusercontent.com/docker/docker-install/master/install.sh

# 3. Xem file đã download
ls -lh install.sh

# 4. Xem ports đang mở
sudo netstat -tulpn

# 5. Xem IP
ip addr
```

---

### 1.8: System Information

**Mục tiêu:** Kiểm tra thông tin server

```bash
# OS info
lsb_release -a
cat /etc/os-release
uname -a

# CPU info
nproc  # Số cores
lscpu

# Memory info
free -h
cat /proc/meminfo

# Disk info
df -h
du -sh folder/  # Disk usage của folder

# Uptime
uptime
```

**Bài Tập:**

```bash
# 1. Xem OS
lsb_release -a

# 2. Xem CPU
nproc
lscpu | grep "Model name"

# 3. Xem RAM
free -h

# 4. Xem disk
df -h

# 5. Xem uptime
uptime
```

---

### 📋 CHECKLIST LINUX COMMANDS

- [ ] Điều hướng: cd, ls, pwd, mkdir, rm
- [ ] File operations: cat, nano, cp, mv
- [ ] Permissions: chmod, chown
- [ ] Package management: apt update, apt install
- [ ] Process management: ps, top, kill
- [ ] Network: ping, curl, wget
- [ ] System info: free, df, uname

---

## 🔐 CHƯƠNG 2: SSH & REMOTE ACCESS

### Tại Sao Học SSH?

```
✅ Kết nối vào server từ xa
✅ Bảo mật hơn password
✅ Mọi cloud đều dùng SSH
✅ Không biết SSH = không làm được DevOps
```

---

### 2.1: SSH Basics

**Mục tiêu:** Kết nối vào server

```bash
# Cú pháp cơ bản
ssh user@hostname

# Ví dụ
ssh ubuntu@192.168.1.100
ssh root@example.com

# Với port khác (mặc định 22)
ssh -p 2222 user@hostname

# Với SSH key
ssh -i /path/to/key.pem user@hostname
ssh -i ~/.ssh/my-key.pem ubuntu@54.123.456.789
```

---

### 2.2: SSH Keys

**Mục tiêu:** Tạo và quản lý SSH keys

```bash
# Tạo SSH key pair
ssh-keygen -t rsa -b 4096 -C "your@email.com"
# Output:
# - Private key: ~/.ssh/id_rsa
# - Public key: ~/.ssh/id_rsa.pub

# Xem public key
cat ~/.ssh/id_rsa.pub

# Copy public key lên server
ssh-copy-id user@hostname

# Hoặc thủ công
cat ~/.ssh/id_rsa.pub | ssh user@hostname "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"
```

---

### 2.3: SSH Config

**Mục tiêu:** Đơn giản hóa SSH commands

```bash
# Tạo SSH config
nano ~/.ssh/config
```

**Nội dung:**

```
Host aws-server
    HostName 54.123.456.789
    User ubuntu
    IdentityFile ~/.ssh/aws-key.pem

Host oracle-server
    HostName 123.456.789.012
    User ubuntu
    IdentityFile ~/.ssh/oracle-key.key

Host my-vps
    HostName example.com
    User root
    Port 2222
```

**Sử dụng:**

```bash
# Thay vì:
ssh -i ~/.ssh/aws-key.pem ubuntu@54.123.456.789

# Chỉ cần:
ssh aws-server
```

---

### 2.4: SCP (Secure Copy)

**Mục tiêu:** Copy files giữa local và server

```bash
# Copy file từ local lên server
scp file.txt user@hostname:/path/to/destination/
scp -i key.pem file.txt ubuntu@54.123.456.789:~/

# Copy folder
scp -r folder/ user@hostname:/path/to/destination/

# Copy từ server về local
scp user@hostname:/path/to/file.txt ./
scp -i key.pem ubuntu@54.123.456.789:~/file.txt ./
```

---

### 📋 CHECKLIST SSH

- [ ] SSH vào server: ssh user@hostname
- [ ] SSH với key: ssh -i key.pem user@hostname
- [ ] Tạo SSH key: ssh-keygen
- [ ] Setup SSH config
- [ ] Copy files: scp

---

## 🐳 CHƯƠNG 3: DOCKER & CONTAINERIZATION

### Tại Sao Học Docker?

```
✅ Deploy nhất quán (local = production)
✅ Không lo "works on my machine"
✅ Dễ scale
✅ Industry standard
✅ Mọi cloud đều support Docker
```

---

### 3.1: Docker Concepts

**Container vs Virtual Machine:**

```
Virtual Machine:
├─ Hardware
├─ Hypervisor
├─ Guest OS (Ubuntu)
├─ Libraries
└─ App

Container:
├─ Hardware
├─ Host OS
├─ Docker Engine
└─ Container (App + Libraries)

→ Container nhẹ hơn, nhanh hơn!
```

---

### 3.2: Docker Commands

```bash
# Xem version
docker --version

# Xem images
docker images

# Pull image
docker pull ubuntu:22.04
docker pull nginx:latest

# Xem containers đang chạy
docker ps
docker ps -a  # Tất cả (kể cả stopped)

# Chạy container
docker run ubuntu:22.04
docker run -d nginx  # Detached mode (background)
docker run -d -p 8080:80 nginx  # Map port
docker run -d -p 8080:80 --name my-nginx nginx  # Đặt tên

# Stop container
docker stop container_id
docker stop my-nginx

# Start container
docker start container_id

# Restart container
docker restart container_id

# Xóa container
docker rm container_id
docker rm -f container_id  # Force remove

# Xóa image
docker rmi image_id

# Xem logs
docker logs container_id
docker logs -f container_id  # Follow (real-time)

# Vào trong container
docker exec -it container_id bash
docker exec -it my-nginx bash

# Xem resource usage
docker stats
```

---

### 3.3: Dockerfile

**Mục tiêu:** Tạo custom Docker image

```dockerfile
# Ví dụ: Dockerfile cho Spring Boot app
FROM maven:3.9.6-eclipse-temurin-17 AS build

WORKDIR /app

COPY pom.xml .
RUN mvn dependency:go-offline -B

COPY src ./src
RUN mvn clean package -DskipTests

FROM eclipse-temurin:17-jre-alpine

WORKDIR /app

COPY --from=build /app/target/*.jar app.jar

EXPOSE 8080

ENTRYPOINT ["java", "-jar", "app.jar"]
```

**Build image:**

```bash
# Build
docker build -t my-app:latest .

# Run
docker run -d -p 8080:8080 my-app:latest
```

---

### 3.4: Docker Compose

**Mục tiêu:** Quản lý nhiều containers

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SPRING_DATASOURCE_URL=jdbc:postgresql://db:5432/mydb
      - SPRING_DATASOURCE_USERNAME=postgres
      - SPRING_DATASOURCE_PASSWORD=password
    depends_on:
      - db

  db:
    image: postgres:15
    environment:
      - POSTGRES_DB=mydb
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
    volumes:
      - db-data:/var/lib/postgresql/data

volumes:
  db-data:
```

**Commands:**

```bash
# Start
docker-compose up -d

# Stop
docker-compose stop

# Restart
docker-compose restart

# Xem logs
docker-compose logs -f

# Xóa
docker-compose down
docker-compose down -v  # Xóa cả volumes
```

---

### 📋 CHECKLIST DOCKER

- [ ] Hiểu Container vs VM
- [ ] docker ps, docker images
- [ ] docker run, docker stop
- [ ] docker logs, docker exec
- [ ] Viết Dockerfile
- [ ] docker build, docker run
- [ ] docker-compose up, docker-compose logs

---

## 🔧 CHƯƠNG 4: GIT & VERSION CONTROL

### Tại Sao Học Git?

```
✅ Quản lý code
✅ Làm việc nhóm
✅ Deploy từ GitHub
✅ Industry standard
```

---

### 4.1: Git Basics

```bash
# Config Git
git config --global user.name "Your Name"
git config --global user.email "your@email.com"

# Init repository
git init

# Clone repository
git clone https://github.com/user/repo.git

# Check status
git status

# Add files
git add file.txt
git add .  # Add tất cả

# Commit
git commit -m "Add new feature"

# Push
git push origin main

# Pull
git pull origin main

# Xem history
git log
git log --oneline
```

---

### 4.2: Branches

```bash
# Tạo branch
git branch feature-login

# Chuyển branch
git checkout feature-login
git switch feature-login  # Modern command

# Tạo và chuyển branch
git checkout -b feature-login

# Xem branches
git branch
git branch -a  # Tất cả (kể cả remote)

# Merge branch
git checkout main
git merge feature-login

# Xóa branch
git branch -d feature-login
```

---

### 📋 CHECKLIST GIT

- [ ] git clone, git pull
- [ ] git add, git commit, git push
- [ ] git branch, git checkout
- [ ] git merge
- [ ] git log, git status

---

## 🌐 CHƯƠNG 5: NETWORKING BASICS

### 5.1: IP Address

```
Public IP: Địa chỉ trên internet
→ 54.123.456.789 (AWS)
→ 123.456.789.012 (Oracle)

Private IP: Địa chỉ trong mạng nội bộ
→ 192.168.1.100
→ 10.0.0.5
```

---

### 5.2: Ports

```
Port: Cổng để truy cập service

Common ports:
- 22: SSH
- 80: HTTP
- 443: HTTPS
- 3306: MySQL
- 5432: PostgreSQL
- 8080: Spring Boot (thường dùng)
```

---

### 5.3: Firewall

```
Firewall: Tường lửa kiểm soát traffic

Inbound Rules: Traffic vào server
→ Mở port 22 (SSH)
→ Mở port 8080 (App)

Outbound Rules: Traffic ra khỏi server
→ Thường allow all
```

---

### 📋 CHECKLIST NETWORKING

- [ ] Hiểu Public IP vs Private IP
- [ ] Hiểu Ports (22, 80, 443, 8080)
- [ ] Hiểu Firewall (Inbound/Outbound)

---

## 🔧 CHƯƠNG 6: TROUBLESHOOTING

### 6.1: Xem Logs

```bash
# App logs (Docker)
docker logs container_id
docker logs -f container_id
docker-compose logs -f

# System logs
sudo journalctl -f
sudo journalctl -u nginx

# File logs
tail -f /var/log/nginx/access.log
tail -f /var/log/nginx/error.log
```

---

### 6.2: Debug Network

```bash
# Test connectivity
ping google.com
ping 8.8.8.8

# Test port
telnet hostname 8080
nc -zv hostname 8080

# Test API
curl http://localhost:8080/api/profiles
curl -v http://localhost:8080/api/profiles  # Verbose

# Xem ports đang mở
sudo netstat -tulpn | grep 8080
sudo ss -tulpn | grep 8080
```

---

### 6.3: Debug Process

```bash
# Xem process
ps aux | grep java
ps aux | grep nginx

# Xem resource usage
top
htop
docker stats

# Kill process
kill PID
kill -9 PID
```

---

### 📋 CHECKLIST TROUBLESHOOTING

- [ ] Xem logs: docker logs, journalctl, tail
- [ ] Test network: ping, curl, telnet
- [ ] Debug process: ps, top, kill

---

## 🎯 PHẦN 2: CHỌN CLOUD

> **Bây giờ bạn đã vững base, chọn cloud nào cũng dễ!**

---

## 📊 SO SÁNH CLOUDS

| Cloud | Chi Phí | RAM | Phổ Biến | Quan Trọng Cho CV |
|-------|---------|-----|----------|-------------------|
| **AWS** | $10-15/tháng (sau 12 tháng) | 1GB | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Oracle** | $0 mãi mãi | 1-6GB | ⭐⭐⭐ | ⭐⭐⭐ |
| **GCP** | $10-20/tháng | 0.6GB | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Railway** | $5/tháng | 512MB | ⭐⭐ | ⭐⭐ |

---

## 🎯 CHỌN CLOUD NÀO?

### Cho Learning:

```
Oracle Cloud
→ Free mãi mãi
→ RAM nhiều
→ Đủ để học
```

### Cho Career:

```
AWS
→ 90% công ty dùng
→ Quan trọng nhất cho CV
→ Interview luôn hỏi
```

### Cho Production:

```
AWS (primary) + Oracle (backup)
→ AWS: Production
→ Oracle: Dev/Staging
```

---

## 🚀 APPLY KIẾN THỨC BASE VÀO CLOUD

### Bạn Đã Biết:

```
✅ Linux commands
✅ SSH
✅ Docker
✅ Git
✅ Networking
✅ Troubleshooting
```

### Chuyển Sang Cloud Chỉ Cần:

```
1. Tạo VM (5 phút)
   → Click button trên UI

2. SSH vào server (1 phút)
   → ssh -i key.pem ubuntu@ip

3. Apply kiến thức base (10 phút)
   → apt update
   → Cài Docker
   → Clone code
   → docker-compose up

Tổng: 15-20 phút!
```

---

## 📚 TÀI LIỆU TIẾP THEO

### Đã Học Xong Base?

Chọn cloud và đọc guide:

1. **AWS:** `docs/AWS-EC2-TU-SO-0.md`
2. **Oracle Cloud:** `docs/ORACLE-CLOUD-TU-SO-0.md`
3. **So sánh:** `docs/LO-TRINH-CHON-CLOUD.md`

---

## 🎓 KIỂM TRA KIẾN THỨC

### Bạn Đã Sẵn Sàng Chọn Cloud Khi:

- [ ] Biết 20+ Linux commands
- [ ] SSH vào server được
- [ ] Chạy Docker containers được
- [ ] Clone code từ GitHub được
- [ ] Xem logs và debug được
- [ ] Hiểu IP, Port, Firewall

**Nếu tất cả ✅ → Chọn cloud bất kỳ, bạn sẽ thành công!**

---

## 💡 LỜI KẾT

```
"Giống như học C trước Java"

C dạy bạn:
→ Pointers, Memory, How computer works

Java dạy bạn:
→ OOP, Framework, How to build apps

Tương tự:

Linux/Docker dạy bạn:
→ Commands, SSH, How server works

AWS/Oracle dạy bạn:
→ UI, Console, How to use cloud

Học nền tảng vững → Công cụ nào cũng dễ!
```

---

**Chúc bạn thành công! 🚀**

*"Master the fundamentals, the tools will follow!"*
