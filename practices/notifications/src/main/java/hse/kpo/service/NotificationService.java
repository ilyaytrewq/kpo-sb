package hse.kpo.service;

import java.time.LocalDateTime;
import java.util.List;

import hse.kpo.domains.Notification;
import hse.kpo.dto.requests.NotificationRequest;
import hse.kpo.dto.responses.NotificationResponse;
import hse.kpo.repository.NotificationRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
@Slf4j
public class NotificationService {

    private final NotificationRepository notificationRepository;

    public NotificationResponse sendNotification(NotificationRequest request) {
        Notification notification = new Notification();
        notification.setRecipient(request.recipient());
        notification.setMessage(request.message());
        notification.setType(request.type() != null ? request.type() : "EMAIL");

        Notification saved = notificationRepository.save(notification);

        // Асинхронная обработка
        processNotificationAsync(saved);

        return new NotificationResponse(
                saved.getId(),
                saved.getStatus(),
                "Уведомление принято в обработку"
        );
    }

    public void processNotificationAsync(Notification notification) {
        try {
            log.info("Отправка уведомления для: {} - {}",
                    notification.getRecipient(), notification.getMessage());

            // Имитация отправки
            Thread.sleep(2000);

            notification.setStatus("SENT");
            notification.setSentAt(LocalDateTime.now());
            notificationRepository.save(notification);

            log.info("Уведомление успешно отправлено: {}", notification.getId());

        } catch (Exception e) {
            log.error("Ошибка отправки уведомления: {}", e.getMessage());
            handleNotificationFailure(notification);
        }
    }

    private void handleNotificationFailure(Notification notification) {
        if (notification.getRetryCount() < 3) {
            notification.setRetryCount(notification.getRetryCount() + 1);
            notificationRepository.save(notification);

            // Повторная попытка
            try {
                Thread.sleep(5000);
                processNotificationAsync(notification);
            } catch (InterruptedException ie) {
                Thread.currentThread().interrupt();
            }
        } else {
            notification.setStatus("FAILED");
            notificationRepository.save(notification);
        }
    }

    public NotificationResponse getNotificationStatus(Long id) {
        return notificationRepository.findById(id)
                .map(notification -> new NotificationResponse(
                        notification.getId(),
                        notification.getStatus(),
                        "Статус: " + notification.getStatus()
                ))
                .orElse(new NotificationResponse(
                        id,
                        "NOT_FOUND",
                        "Уведомление не найдено"
                ));
    }

    public List<Notification> getAllNotifications() {
        return notificationRepository.findAll();
    }
}