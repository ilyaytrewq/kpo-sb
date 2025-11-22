package hse.kpo.dto.responses;

import lombok.Builder;

@Builder
public record NotificationResponse(
        Long notificationId,
        String status,
        String message
) {}