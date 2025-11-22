package hse.kpo.config.swagger;

import hse.kpo.integration.notification.NotificationIntegrationProperties;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.MediaType;
import org.springframework.web.client.RestClient;

@Configuration
@EnableConfigurationProperties(NotificationIntegrationProperties.class)
public class RestClientConfig {
    @Bean
    public RestClient notificationRestClient(NotificationIntegrationProperties properties) {
        return RestClient.builder()
                .baseUrl(properties.url())
                .defaultHeader("Content-Type", "application/json")
                .defaultHeader("Accept", "application/json")
                .build();
    }
}
