package hse.kpo.clients;

import hse.kpo.config.NotificationIntegrationProperty;
import hse.kpo.dto.request.NotificationRequest;
import hse.kpo.dto.response.NotificationResponse;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;

@Component
@RequiredArgsConstructor
@Slf4j
public class NotificationClient {

    private final NotificationIntegrationProperty properties;

    private final RestClient notificationRestClient;

    public NotificationResponse sendNotification(NotificationRequest request) {
        try {
            ResponseEntity<NotificationResponse> response = notificationRestClient
                    .post()
                    .uri(properties.getAllEndpoint())
                    .body(request)
                    .retrieve()
                    .toEntity(NotificationResponse.class);

            log.info("Уведомление отправлено успешно: {}", response.getBody());
            return response.getBody();

        } catch (Exception e) {
            log.error("Ошибка при отправке уведомления: {}", e.getMessage());
            return new NotificationResponse(
                    null,
                    "ERROR",
                    "Не удалось отправить уведомление: " + e.getMessage()
            );
        }
    }

}