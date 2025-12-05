import json
import time
import webbrowser
import numpy as np
import paho.mqtt.client as mqtt
from PyQt5.uic import loadUi
from PyQt5.QtCore import QTimer
from PyQt5.QtGui import QPixmap, QImage
from PyQt5.QtWidgets import QMainWindow, QMessageBox

from src.services.yolo_service import prediction
from src.ultis.get_resource_path import ResourcePath
from src.services.llm_service.gemini import GeminiChat
from src.services.camera_service.camera import CameraThread

class App(QMainWindow):
    def __init__(self):
        super().__init__()
        self.setFixedSize(964, 582)
        self.ui = loadUi(ResourcePath.UI_PATH, self)

        self.ui.result_infor.setText("")
        self.ui.predicted_res.setText("")
        self.ui.predicting_btn.setText("Chat với Chatbot")
        self.ui.moreInfor_btn.clicked.connect(self.more_info)
        self.ui.predicting_btn.clicked.connect(self.suggestion)

        # Lưu frame gốc từ camera
        self.current_frame = None

        # MQTT Configuration
        self.mqtt_broker = "YOUR_MQTT_BROKER_ADDRESS"  # Thay đổi địa chỉ broker
        self.mqtt_port = 1883
        self.user_id = "YOUR_USER_ID"  # Thay đổi user ID của bạn
        self.mcu_code = "YOUR_MCU_CODE"  # Thay đổi MCU code của bạn
        self.mqtt_topic = f"user/{self.user_id}/mcu/{self.mcu_code}/alert"

        # Initialize MQTT Client
        self.mqtt_client = mqtt.Client()
        self.mqtt_client.on_connect = self.on_mqtt_connect
        self.mqtt_client.on_disconnect = self.on_mqtt_disconnect

        try:
            self.mqtt_client.connect(self.mqtt_broker, self.mqtt_port, 60)
            self.mqtt_client.loop_start()
        except Exception as e:
            print(f"Không thể kết nối MQTT broker: {e}")

        self.camera_thread = CameraThread()
        self.camera_thread.frame_signal.connect(self.update_frame)
        self.camera_thread.start()

        self.auto_predict_timer = QTimer(self)
        self.auto_predict_timer.timeout.connect(self.predict)
        self.auto_predict_timer.start(15000)

    def on_mqtt_connect(self, client, userdata, flags, rc):
        if rc == 0:
            print("Đã kết nối MQTT broker thành công")
        else:
            print(f"Kết nối MQTT thất bại với code: {rc}")

    def on_mqtt_disconnect(self, client, userdata, rc):
        print("Đã ngắt kết nối MQTT broker")

    def send_mqtt_alert(self, disease_name, severity="warning"):
        """
        Gửi cảnh báo qua MQTT
        """
        try:
            alert_payload = {
                "mcu_code": self.mcu_code,
                "title": "Phát hiện bệnh cây",
                "message": f"Hệ thống phát hiện: {disease_name}",
                "severity": severity,
                "time": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
            }

            mqtt_message = {
                "topic": self.mqtt_topic,
                "payload": alert_payload,
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
            }

            message_json = json.dumps(mqtt_message)
            result = self.mqtt_client.publish(self.mqtt_topic, message_json, qos=1)

            if result.rc == mqtt.MQTT_ERR_SUCCESS:
                print(f"Đã gửi cảnh báo MQTT: {disease_name}")
            else:
                print(f"Lỗi gửi MQTT: {result.rc}")

        except Exception as e:
            print(f"Lỗi khi gửi cảnh báo MQTT: {e}")

    def update_frame(self, frame):
        """ Cập nhật ảnh từ camera lên giao diện """
        # Lưu frame gốc (QImage) để sử dụng cho prediction
        self.current_frame = frame
        self.ui.realtime_img.setPixmap(QPixmap.fromImage(frame))

    def qimage_to_cv2(self, qimage):
        """
        Chuyển đổi QImage sang numpy array (OpenCV format)
        """
        try:
            # Chuyển QImage sang RGB format nếu cần
            if qimage.format() != QImage.Format_RGB888:
                qimage = qimage.convertToFormat(QImage.Format_RGB888)

            # Lấy kích thước
            width = qimage.width()
            height = qimage.height()

            # Chuyển đổi sang numpy array
            ptr = qimage.bits()
            ptr.setsize(height * width * 3)
            arr = np.array(ptr).reshape(height, width, 3)

            # Chuyển từ RGB sang BGR (OpenCV format)
            arr = arr[:, :, ::-1].copy()

            return arr
        except Exception as e:
            print(f"Lỗi chuyển đổi QImage sang CV2: {e}")
            return None

    def cv2_to_qimage(self, cv_img):
        """
        Chuyển đổi numpy array (OpenCV format) sang QImage
        """
        try:
            height, width, channel = cv_img.shape
            bytes_per_line = 3 * width

            # Chuyển từ BGR sang RGB
            rgb_image = cv_img[:, :, ::-1].copy()

            q_image = QImage(rgb_image.data, width, height, bytes_per_line, QImage.Format_RGB888)
            return q_image
        except Exception as e:
            print(f"Lỗi chuyển đổi CV2 sang QImage: {e}")
            return None

    def predict(self):
        """ Dự đoán bệnh cây từ ảnh camera """
        if self.current_frame is None:
            print("Không có frame để dự đoán")
            return

        try:
            # Chuyển đổi QImage sang OpenCV format
            cv_image = self.qimage_to_cv2(self.current_frame)

            if cv_image is None:
                print("Lỗi chuyển đổi ảnh")
                return

            print(f"Kích thước ảnh đầu vào: {cv_image.shape}")

            # Gọi hàm dự đoán
            res_name, res_description, res_image = prediction.yolo_prediction(cv_image)

            # Chuyển đổi kết quả về QImage để hiển thị
            result_qimage = self.cv2_to_qimage(res_image)

            if result_qimage is not None:
                res_pixmap = QPixmap.fromImage(result_qimage)
                res_img_scaled = res_pixmap.scaled(300, 250)
                self.ui.predicted_img.setPixmap(res_img_scaled)

            # Gửi cảnh báo qua MQTT nếu phát hiện bệnh
            if res_name != "Cây khỏe mạnh" and res_name != "Lỗi dự đoán":
                # Xác định mức độ nghiêm trọng dựa vào tên bệnh
                severity = "warning"  # Mặc định là warning
                # Bạn có thể tùy chỉnh severity dựa vào loại bệnh
                if "nặng" in res_name.lower() or "nghiêm trọng" in res_name.lower():
                    severity = "error"
                elif "nhẹ" in res_name.lower():
                    severity = "info"

                self.send_mqtt_alert(res_name, severity)

            # Cập nhật giao diện
            self.ui.predicted_res.setText(res_name)
            if res_name == "Cây khỏe mạnh":
                self.ui.result_infor.setText("Bạn không cần phải lo lắng, cây của bạn đang khỏe mạnh!")
            elif res_name == "Lỗi dự đoán":
                self.ui.result_infor.setText("Có lỗi xảy ra trong quá trình dự đoán. Vui lòng thử lại.")
            else:
                self.ui.result_infor.setText(f"Mô tả tình trạng {res_name}: {res_description}")

        except Exception as e:
            print(f"Lỗi trong quá trình dự đoán: {e}")
            self.ui.predicted_res.setText("Lỗi dự đoán")
            self.ui.result_infor.setText("Có lỗi xảy ra trong quá trình dự đoán. Vui lòng thử lại.")

    def more_info(self):
        """ Mở Google tìm kiếm thông tin về bệnh cây đã dự đoán """
        predicted_text = self.ui.predicted_res.text().strip()
        if predicted_text != "" and predicted_text != "Lỗi dự đoán":
            search_query = predicted_text.replace(" ", "+")
            webbrowser.open(f"https://www.google.com/search?q={search_query}")
        else:
            QMessageBox.information(self, "Thông báo", "Chưa có kết quả dự đoán để tìm kiếm.")

    def suggestion(self):
        """ Mở chức năng chatbot và hiển thị kết quả qua popup """
        predicted_text = self.ui.predicted_res.text().strip()

        if predicted_text == "" or predicted_text == "Lỗi dự đoán":
            QMessageBox.information(self, "Thông báo", "Bạn chưa có kết quả dự đoán hợp lệ để chat.")
        else:
            try:
                response = GeminiChat(predicted_text)
                if response:
                    QMessageBox.information(self, "Kết quả từ Chatbot", response)
                else:
                    QMessageBox.information(self, "Thông báo", "Không có phản hồi từ Chatbot.")
            except Exception as e:
                print(f"Error when call API: {e}")
                QMessageBox.warning(self, "Lỗi", f"Không thể kết nối với Chatbot: {str(e)}")

    def closeEvent(self, event):
        """ Xử lý khi đóng ứng dụng """
        if hasattr(self, 'camera_thread'):
            self.camera_thread.stop()
            self.camera_thread.wait()

        # Ngắt kết nối MQTT
        if hasattr(self, 'mqtt_client'):
            self.mqtt_client.loop_stop()
            self.mqtt_client.disconnect()

        event.accept()
