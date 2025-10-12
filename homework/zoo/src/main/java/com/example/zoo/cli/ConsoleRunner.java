package com.example.zoo.cli;

import com.example.zoo.domain.animal.*;
import com.example.zoo.domain.thing.*;
import com.example.zoo.service.ReportService;
import com.example.zoo.service.ZooService;
import com.example.zoo.service.dto.ZooSummaryDto;
import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

import java.util.Scanner;


@Component
@TestPropertySource(properties = "app.console.enabled=false")
public class ConsoleRunner implements CommandLineRunner {
    private final ZooService zoo;
    private final ReportService report;

    public ConsoleRunner(ZooService zoo, ReportService report) {
        this.zoo = zoo;
        this.report = report;
    }

    @Override
    public void run(String... args) {
        Scanner scanner = new Scanner(System.in);
        if (!scanner.hasNextLine()) {
            // среда без пользовательского ввода (например, тесты/CI) — просто выходим
            System.out.println("No interactive input. Skipping console mode.");
            return;
        }

        boolean running = true;
        System.out.println("================= ЗООПАРК =================");
        while (running) {
            System.out.println("""
                    Выберите действие:
                    1. Добавить животное
                    2. Добавить вещь
                    3. Показать отчёт
                    0. Выход
                    """);

            System.out.print("Введите номер действия: ");
            if (!scanner.hasNextLine()) break;
            String choice = scanner.nextLine().trim();

            switch (choice) {
                case "1" -> {/* addAnimal(scanner); */}
                case "2" -> {/* addThing(scanner); */}
                case "3" -> {/* showReport(); */}
                case "0" -> running = false;
                default -> System.out.println("Неверный выбор.");
            }
        }
        System.out.println("Программа завершена.");
    }
}
