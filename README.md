# Lưu data

## Server lobby

> Một server duy nhất, chứa toàn bộ thông tin người dùng, và trỏ đến các server game, điều hướng người chơi vào server game

### Dữ liệu

- Database: map[username:string->Account]
  - Account:
    - Username: string
    - Session: \*KCPConnection
    - GameServerAddress: \*string (pointer = nullable)
- GameServerAddresses: string\[]
- Rooms: Room[]
  - Room:
    - CurrentPlayers: int
    - MaxPlayers: int
    - GameServerAddress: string
    - State: RoomState

### Cổng:

- 19000: nghe server game
- 19001: nghe client

## Server game

> Nhiều server đặt ở các region khác nhau (mỗi region có thể nhiều server game -> thành cụm server game) để xử lý logic các phòng game

### Dữ liệu

- LobbyServerSession: KCPConnection
- Rooms: Room\[]
- Room:
  - Map: GameMap
  - GameMap:
    - Dữ liệu map (size, tile, ...)
    - Store là toàn bộ vật phẩm game (mũ, áo, quần, súng)
  - Players: map[username:string->Player]
    - Player
      - Session: KCPConnection
      - Dữ liệu người chơi trong game (vị trí, máu, vật phẩm hiện có,...)
- Players: map[username:string->*Room]

### Cổng:

-   19002: nghe server lobby
-   19003: nghe client

# Giao thức game PWNBG

## Packet dạng chung (cho cả server lobby (SL), server game (SG) và client(C)):

-   **Byte 0:** Loại gói tin
-   **Byte 1:** Loại gói tin chi tiết hơn
-   **Các byte sau:** Thông tin chi tiết

## C khởi tạo kết nối tới SL (byte 0 = 1):

-   **Byte 1:** Loại gói tin

### C gửi lệnh đăng nhập

-   **Byte 1** = 1
-   **Byte 2+**: Tên người chơi (string)

### SL gửi lệnh đăng nhập thành công

-   **Byte 1** = 2

### SL gửi lệnh đăng nhập thất bại

-   **Byte 1** = 3

## SG khởi tạo kết nối với SL (byte 0 = 2):

-   **Byte 1:** Loại gói tin

### SG gửi lệnh muốn thêm vào SL, với các cài đặt hiện tại của SL

-   **Byte 1** = 1
-   **Byte 2-3**: Tần suất gửi gói tin có tần suất cố định của client/server (gói/giây) (bội số của 2, mặc định là 64)
-   **Byte 4-5**: Số lượng room tối đa (bội số của 2, mặc định là 4)
-   **Byte 6-7**: Số lượng player tối đa (mặc định là 20)

### SL gửi thông báo xác nhận thành công

-   **Byte 1** = 2

### SL gửi thông báo xác nhận thất bại

-   **Byte 1** = 3

## SL gửi địa chỉ để C kết nối tới SG (byte 0 = 3):

-   **Byte 1:** Loại gói tin

### SL gửi địa chỉ SG gần nhất cho C

-   **Byte 1** = 1
-   **Byte 2+**: Địa chỉ SG. VD: "192.168.1.3"

### SL bắt C phải ngắt kết nối tới SG

-   **Byte 1** = 2

## SG trao đổi thông tin các phòng với SL:

-   **Byte 1:** Loại gói tin

### SG gửi tình trạng của phòng (5s/lần)

- **Byte 1** = 1
- Nhóm 5 byte từ byte số 2:
  - **Byte 0-1**: Số người chơi hiện tại
  - **Byte 2-3**: Số người chơi tối đa
  - **Byte 4**: Trạng thái game (0 = đang chờ, 1 = đang chơi)

## C kết nối tới SG (byte 0 = 5):

-   **Byte 1:** Loại gói tin

### C gửi thông báo muốn vào phòng

- **Byte 1:** = 1
- **Byte 2+:** Tên đăng nhập

### SG gửi thông báo xác nhận thành công, và dữ liệu trạng thái game hiện tại

- **Byte 1** = 2
- **Byte 2** = 1 nếu gói tin có kèm dữ liệu hiện tại, 0 nếu không gửi
- **Byte 3-4:** Toạ độ x hiện tại của người chơi
- **Byte 5-6:** Toạ độ y hiện tại của người chơi
- **Byte 7-8:** Máu hiện tại của người chơi
- **Byte 9-10:** ID súng đang trang bị của người chơi
- **Byte 11-12:** ID mũ đang trang bị của người chơi
- **Byte 13-14:** ID áo đang trang bị của người chơi
- **Byte 15-16:** ID quần đang trang bị của người chơi
- **Các byte sau:** //TODO Tính sau

### SG gửi thông báo xác nhận thất bại

-   **Byte 1** = 3

## Trong game (byte 0 = 6):

-   **Byte 1:** Loại cập nhật trong game
-   **Các byte sau:** Thông tin chi tiết

### C: Cập nhật vị trí người chơi

-   **Byte 1** = 1
-   **Byte 2-3:** Toạ độ x
-   **Byte 4-5:** Toạ độ y

### SG: Cập nhật item (trang bị mới nhặt được) của người chơi

-   **Byte 1** = 2
-   **Byte 2-3:** ID item

### SG: Cập nhật máu người chơi

- **Byte 1** = 3
- **Byte 2-3:** Máu bị trừ
- **Byte 4-5:** Máu hiện tại

### SG: Thông báo trò chơi kết thúc

- **Byte 1** = 4