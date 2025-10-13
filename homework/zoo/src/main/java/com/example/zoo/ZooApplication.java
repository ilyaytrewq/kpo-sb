package com.example.zoo;

import com.example.zoo.domain.animal.*;
import com.example.zoo.domain.thing.*;
import com.example.zoo.service.ReportService;
import com.example.zoo.service.ZooService;
import com.example.zoo.service.dto.ZooSummaryDto;
import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;

import java.util.Locale;
import java.util.Scanner;

@SpringBootApplication
public class ZooApplication {

    public static void main(String[] args) {
        SpringApplication.run(ZooApplication.class, args);
    }

    @Bean
    public CommandLineRunner consoleRunner(ZooService zoo, ReportService report) {
        return args -> {
            Scanner scanner = new Scanner(System.in);

            boolean running = true;
            while (running) {
                System.out.println("""
                        ============================
                        Добро пожаловать в зоопарк!
                        Выберите действие:
                        1. Добавить животное
                        2. Добавить вещь
                        3. Показать отчёт
                        0. Выход
                        ============================
                        """);
                System.out.print("Введите номер действия: ");
                if (!scanner.hasNextLine()) break;
                String choice = scanner.nextLine().trim();

                switch (choice) {
                    case "1" -> addAnimal(scanner, zoo);
                    case "2" -> addThing(scanner, zoo);
                    case "3" -> showReport(report);
                    case "0" -> running = false;
                    default -> System.out.println("Неверный выбор, попробуйте снова");
                }
            }
            System.out.println("Программа завершена.");
        };
    }

    private static void addAnimal(Scanner scanner, ZooService zoo) {
        System.out.print("Введите тип животного (Rabbit/Monkey/Tiger/Wolf): ");
        if (!scanner.hasNextLine()) return;
        String type = scanner.nextLine().trim().toLowerCase(Locale.ROOT);

        System.out.print("Введите имя: ");
        if (!scanner.hasNextLine()) return;
        String name = scanner.nextLine().trim();

        Integer food = readInt(scanner, "Введите кг корма в день (целое): ");
        if (food == null) return;

        Integer id = readInt(scanner, "Введите инвентарный номер (целое): ");
        if (id == null) return;

        Animal animal = null;

        switch (type) {
            case "rabbit", "monkey" -> {
                Integer kindness = readInt(scanner, "Введите доброту (0–10): ");
                if (kindness == null) return;
                animal = type.equals("rabbit")
                        ? new Rabbit(name, food, id, kindness)
                        : new Monkey(name, food, id, kindness);
            }
            case "tiger", "wolf" -> {
                Integer danger = readInt(scanner, "Введите уровень опасности (0–10): ");
                if (danger == null) return;
                animal = type.equals("tiger")
                        ? new Tiger(name, food, id, danger)
                        : new Wolf(name, food, id, danger);
            }
            default -> {
                System.out.println("Неизвестный тип животного.");
                return;
            }
        }

        boolean accepted = zoo.admit(animal);
        System.out.println(accepted
                ? "Животное принято в зоопарк"
                : "Ветеринар не одобрил животное");
    }

    private static void addThing(Scanner scanner, ZooService zoo) {
        System.out.print("Введите тип вещи (Table/Computer): ");
        if (!scanner.hasNextLine()) return;
        String type = scanner.nextLine().trim().toLowerCase(Locale.ROOT);

        System.out.print("Введите название: ");
        if (!scanner.hasNextLine()) return;
        String name = scanner.nextLine().trim();

        Integer number = readInt(scanner, "Введите инвентарный номер (целое): ");
        if (number == null) return;

        Thing thing = switch (type) {
            case "table" -> new Table(number, name);
            case "computer" -> new Computer(number, name);
            default -> null;
        };

        if (thing == null) {
            System.out.println("Неизвестный тип вещи");
            return;
        }

        zoo.addThing(thing);
        System.out.println("Вещь добавлена в инвентарь");
    }

    private static void showReport(ReportService report) {
        ZooSummaryDto s = report.buildSummary();
        System.out.printf("Животных принято: %d%n", s.animalsCount());
        System.out.printf("Корма нужно в сутки: %d кг%n", s.totalFoodKgPerDay());

        System.out.println("Кандидаты в контактный зоопарк:");
        if (s.interactiveNames().isEmpty()) System.out.println("  — нет подходящих");
        else s.interactiveNames().forEach(n -> System.out.println("  • " + n));

        System.out.println("\nИнвентаризация:");
        s.allInventory().forEach(i -> System.out.printf("  #%d — %s%n", i.number(), i.name()));
        System.out.println("==========================================\n");
    }

    private static Integer readInt(Scanner sc, String prompt) {
        System.out.print(prompt);
        if (!sc.hasNextLine()) return null;
        String s = sc.nextLine().trim();
        try {
            return Integer.parseInt(s);
        } catch (NumberFormatException e) {
            System.out.println("Ожидалось целое число.");
            return null;
        }
    }
}
