package hse.kpo.dto.requests;

import jakarta.validation.constraints.NotBlank;

public record NotificationRequest(
        @NotBlank String recipient,
        @NotBlank String message,
        String type
) {}