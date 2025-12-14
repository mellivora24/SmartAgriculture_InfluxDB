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
from src.services.camera_service.camera import CameraThread


class App(QMainWindow):
    def __init__(self):
        super().__init__()
        self.setFixedSize(964, 582)
        self.ui = loadUi(ResourcePath.UI_PATH, self)

        self.ui.result_infor.setText("")
        self.ui.predicted_res.setText("")
        self.ui.predicting_btn.setText("Kiểm tra cây trồng")
        self.ui.moreInfor_btn.clicked.connect(self.more_info)
        self.ui.predicting_btn.clicked.connect(self.predict)

        # Lưu frame gốc từ camera
        self.current_frame = None

        # MQTT Configuration
        self.mqtt_broker = "localhost"
        self.mqtt_port = 1883
        self.user_id = "admin"  # Thay đổi user ID của bạn
        self.mcu_code = "admin123456"  # Thay đổi MCU code của bạn

        # FIXED: Tạo các topic đúng theo cấu trúc backend
        self.mqtt_topic_alert = f"user/{self.user_id}/mcu/{self.mcu_code}/alert"
        self.mqtt_topic_disease = f"user/{self.user_id}/mcu/{self.mcu_code}/disease_detection"

        # Initialize MQTT Client
        self.mqtt_client = mqtt.Client()
        self.mqtt_client.username_pw_set("admin", "admin123456")
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

    def send_mqtt_disease_detection(self, disease_name, confidence=0.0):
        """
        Gửi thông tin disease detection qua MQTT
        Topic: user/{user_id}/mcu/{mcu_code}/disease_detection
        """
        try:
            # Payload theo cấu trúc DiseaseDetectionPayload trong Go
            disease_payload = {
                "mcu_code": self.mcu_code,
                "disease_name": disease_name,
                "confidence": confidence,
                "detected_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
            }

            # Wrap trong MQTTMessage structure
            mqtt_message = {
                "topic": "disease_detection",
                "payload": disease_payload
            }

            message_json = json.dumps(mqtt_message)
            result = self.mqtt_client.publish(self.mqtt_topic_disease, message_json, qos=1)

            if result.rc == mqtt.MQTT_ERR_SUCCESS:
                print(f"✅ Đã gửi disease detection: {disease_name} (confidence: {confidence:.2f})")
            else:
                print(f"❌ Lỗi gửi disease detection MQTT: {result.rc}")

        except Exception as e:
            print(f"❌ Lỗi khi gửi disease detection MQTT: {e}")

    def send_mqtt_alert(self, disease_name, severity="warning"):
        """
        Gửi cảnh báo qua MQTT (optional - chỉ cho bệnh nghiêm trọng)
        Topic: user/{user_id}/mcu/{mcu_code}/alert
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
                "topic": "alert",
                "payload": alert_payload
            }

            message_json = json.dumps(mqtt_message)
            result = self.mqtt_client.publish(self.mqtt_topic_alert, message_json, qos=1)

            if result.rc == mqtt.MQTT_ERR_SUCCESS:
                print(f"✅ Đã gửi alert MQTT: {disease_name}")
            else:
                print(f"❌ Lỗi gửi alert MQTT: {result.rc}")

        except Exception as e:
            print(f"❌ Lỗi khi gửi alert MQTT: {e}")

    def update_frame(self, frame):
        """ Cập nhật ảnh từ camera lên giao diện """
        self.current_frame = frame
        self.ui.realtime_img.setPixmap(QPixmap.fromImage(frame))

    def qimage_to_cv2(self, qimage):
        """
        Chuyển đổi QImage sang numpy array (OpenCV format)
        """
        try:
            if qimage.format() != QImage.Format_RGB888:
                qimage = qimage.convertToFormat(QImage.Format_RGB888)

            width = qimage.width()
            height = qimage.height()

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
            QMessageBox.warning(self, "Cảnh báo", "Không có hình ảnh từ camera để kiểm tra.")
            return

        try:
            # Chuyển đổi QImage sang OpenCV format
            cv_image = self.qimage_to_cv2(self.current_frame)

            if cv_image is None:
                QMessageBox.warning(self, "Lỗi", "Không thể xử lý hình ảnh.")
                return

            print(f"Kích thước ảnh đầu vào: {cv_image.shape}")

            # Gọi hàm dự đoán
            # IMPORTANT: Bạn cần modify hàm yolo_prediction để trả về confidence score
            res_name, res_description, res_image = prediction.yolo_prediction(cv_image)

            # TODO: Nếu hàm yolo_prediction có thể trả về confidence, sử dụng như sau:
            # res_name, res_description, res_image, confidence = prediction.yolo_prediction(cv_image)
            # Nếu không, dùng giá trị mặc định:
            confidence = 0.85  # Giá trị mặc định, bạn nên lấy từ YOLO model

            # Chuyển đổi kết quả về QImage để hiển thị
            result_qimage = self.cv2_to_qimage(res_image)

            if result_qimage is not None:
                res_pixmap = QPixmap.fromImage(result_qimage)
                res_img_scaled = res_pixmap.scaled(300, 250)
                self.ui.predicted_img.setPixmap(res_img_scaled)

            # FIXED: Luôn gửi disease detection (bao gồm cả trường hợp khỏe mạnh)
            if res_name != "Lỗi dự đoán":
                # Gửi disease detection lên topic disease_detection
                self.send_mqtt_disease_detection(res_name, confidence)

                # Chỉ gửi alert nếu phát hiện bệnh (không phải cây khỏe mạnh)
                if res_name != "Cây khỏe mạnh":
                    # Xác định mức độ nghiêm trọng
                    severity = "warning"  # Mặc định

                    if confidence > 0.8:
                        severity = "error"  # Bệnh nghiêm trọng
                    elif confidence > 0.5:
                        severity = "warning"
                    else:
                        severity = "info"

                    # Override bằng từ khóa trong tên bệnh
                    if "nặng" in res_name.lower() or "nghiêm trọng" in res_name.lower():
                        severity = "error"
                    elif "nhẹ" in res_name.lower():
                        severity = "info"

                    # Gửi alert cho bệnh nghiêm trọng
                    self.send_mqtt_alert(res_name, severity)

            # Cập nhật giao diện
            self.ui.predicted_res.setText(res_name)
            if res_name == "Cây khỏe mạnh":
                self.ui.result_infor.setText("Bạn không cần phải lo lắng, cây của bạn đang khỏe mạnh!")
                QMessageBox.information(self, "Kết quả", "Cây của bạn đang khỏe mạnh!")
            elif res_name == "Lỗi dự đoán":
                self.ui.result_infor.setText("Có lỗi xảy ra trong quá trình dự đoán. Vui lòng thử lại.")
                QMessageBox.warning(self, "Lỗi", "Có lỗi xảy ra trong quá trình dự đoán.")
            else:
                self.ui.result_infor.setText(f"Mô tả tình trạng {res_name}: {res_description}")
                QMessageBox.information(self, "Kết quả kiểm tra", f"Phát hiện: {res_name}\n\n{res_description}")

        except Exception as e:
            print(f"Lỗi trong quá trình dự đoán: {e}")
            self.ui.predicted_res.setText("Lỗi dự đoán")
            self.ui.result_infor.setText("Có lỗi xảy ra trong quá trình dự đoán. Vui lòng thử lại.")
            QMessageBox.critical(self, "Lỗi", f"Có lỗi xảy ra: {str(e)}")

    def more_info(self):
        """ Mở Google tìm kiếm thông tin về bệnh cây đã dự đoán """
        predicted_text = self.ui.predicted_res.text().strip()
        if predicted_text != "" and predicted_text != "Lỗi dự đoán":
            search_query = predicted_text.replace(" ", "+")
            webbrowser.open(f"https://www.google.com/search?q={search_query}")
        else:
            QMessageBox.information(self, "Thông báo", "Chưa có kết quả dự đoán để tìm kiếm.")

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