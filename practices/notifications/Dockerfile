# Build stage
FROM gradle:8.11.1-jdk21 AS build
WORKDIR /app
COPY build.gradle.kts settings.gradle.kts ./
COPY --chown=gradle:gradle . .
RUN gradle dependencies --refresh-dependencies && \
    gradle build -x test --no-daemon


# Run stage
FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=build /app/build/libs/ ./libs/
EXPOSE 8081
CMD ["java", "-jar", "libs/notifications-1.0.jar"]