plugins {
    id 'java'
    id 'org.springframework.boot' version '3.5.6'
    id 'io.spring.dependency-management' version '1.1.7'
    id 'jacoco'
}

group = 'com.example.zoo'
version = '0.0.1-SNAPSHOT'
description = 'zoo'

java {
    toolchain {
        languageVersion = JavaLanguageVersion.of(17)
    }
}

repositories {
    mavenCentral()
}

dependencies {
    implementation 'org.springframework.boot:spring-boot-starter'
    developmentOnly 'org.springframework.boot:spring-boot-devtools'

    testImplementation 'org.springframework.boot:spring-boot-starter-test'
    testRuntimeOnly 'org.junit.platform:junit-platform-launcher'
}

jacoco {
    toolVersion = '0.8.12'
}

test {
    useJUnitPlatform()
    finalizedBy jacocoTestReport
}

jacocoTestReport {
    dependsOn test
            reports {
                html.required = true   // HTML-отчёт в build/reports/jacoco/test/html/index.html
                xml.required  = false
                csv.required  = false
            }
}

bootRun {
    systemProperty 'app.console.enabled', 'true'
}
