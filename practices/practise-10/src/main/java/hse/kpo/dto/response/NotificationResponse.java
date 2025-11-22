package hse.kpo.dto.response;

public record NotificationResponse(
        Long notificationId,
        String status,
        String message
) {}