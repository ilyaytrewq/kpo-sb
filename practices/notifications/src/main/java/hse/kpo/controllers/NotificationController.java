package hse.kpo.controllers;

import hse.kpo.domains.Notification;
import hse.kpo.dto.requests.NotificationRequest;
import hse.kpo.dto.responses.NotificationResponse;
import hse.kpo.service.NotificationService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/notifications")
@RequiredArgsConstructor
@Tag(name = "Уведомления", description = "API для отправки уведомлений")
public class NotificationController {

    private final NotificationService notificationService;

    @PostMapping
    @Operation(summary = "Отправить уведомление")
    public ResponseEntity<NotificationResponse> sendNotification(
            @Valid @RequestBody NotificationRequest request) {
        return ResponseEntity.ok(notificationService.sendNotification(request));
    }

    @GetMapping("/{id}/status")
    @Operation(summary = "Получить статус уведомления")
    public ResponseEntity<NotificationResponse> getNotificationStatus(@PathVariable Long id) {
        return ResponseEntity.ok(notificationService.getNotificationStatus(id));
    }

    @GetMapping
    @Operation(summary = "Получить все уведомления")
    public ResponseEntity<List<Notification>> getAllNotifications() {
        return ResponseEntity.ok(notificationService.getAllNotifications());
    }
}
