package hse.kpo.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "notifications-integration")
public record NotificationIntegrationProperties(String url, String getAllEndpoint) {}