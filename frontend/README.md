# CAU TRUC DU AN
```bash
src/
├── main.tsx
├── App.tsx
├── core/                     # Nền tảng chung
│   ├── api/                  # Axios client, interceptors
│   ├── config/               # Env, constants
│   ├── hooks/                # Custom hook toàn cục
│   ├── types/                # Khai bao kieu du lieu
│   └── utils/                # Helper functions
├── features/                 # Feature riêng từng màn hình / chức năng
│   ├── auth/                 # Đăng nhập / đăng ký
│   │   ├── api/              # call API login/register
│   │   ├── components/       # LoginForm, RegisterForm
│   │   ├── hooks/            # useAuth, useToken
│   │   └── pages/
│   │       ├── LoginPage.tsx
│   │       └── RegisterPage.tsx
│   ├── onboarding/           # Màn hình onboarding
│   │   ├── components/
│   │   └── pages/
│   │       └── OnboardingPage.tsx
│   ├── dashboard/            # Trang tổng quan realtime
│   │   ├── api/              # Gọi farm, MCU, surveyPoint, sensorData
│   │   ├── components/       # CardFarm, CardMCU, SensorChart, CommandCard
│   │   ├── hooks/            # useDashboardData, useRealtimeUpdates
│   │   └── pages/
│   │       └── DashboardPage.tsx
│   ├── surveyPoint/          # Quản lý survey point
│   │   ├── api/
│   │   ├── components/       # SurveyPointForm, SurveyPointTable
│   │   ├── hooks/            # useSurveyPoint
│   │   └── pages/
│   │       └── SurveyPointManagementPage.tsx
│   ├── deviceCommand/        # Lịch sử bật/tắt thiết bị
│   │   ├── api/
│   │   ├── components/       # CommandTable
│   │   ├── hooks/            # useDeviceCommands
│   │   └── pages/
│   │       └── DeviceCommandHistoryPage.tsx
│   ├── sensorData/           # Lịch sử dữ liệu cảm biến
│   │   ├── api/
│   │   ├── components/       # SensorTable, SensorChart
│   │   ├── hooks/            # useSensorData
│   │   └── pages/
│   │       └── SensorDataHistoryPage.tsx
│   ├── account/              # Cài đặt tài khoản
│   │   ├── api/              # Update, delete, change password
│   │   ├── components/       # AccountForm, AvatarUploader
│   │   ├── hooks/            # useAccount
│   │   └── pages/
│   │       └── AccountSettingsPage.tsx
│   └── media/                # Lưu ảnh dùng bên thứ 3
│       ├── api/              # Upload image API / Cloudinary
│       └── hooks/            # useMedia
├── shared/                   # Component dùng chung
│   ├── components/           # Button, Modal, Input, Table...
│   └── ui/                   # Theme, Typography, Layout
├── assets/                   # Ảnh, svg, fonts
└── styles/                   # CSS global / Tailwind
```