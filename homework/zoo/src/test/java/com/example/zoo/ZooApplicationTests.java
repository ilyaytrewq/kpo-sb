package com.example.zoo;

import com.example.zoo.cli.ConsoleRunner;
import com.example.zoo.domain.animal.Monkey;
import com.example.zoo.domain.animal.Rabbit;
import com.example.zoo.domain.animal.Tiger;
import com.example.zoo.domain.animal.Wolf;
import com.example.zoo.domain.thing.Computer;
import com.example.zoo.domain.thing.Table;
import com.example.zoo.containers.impl.AnimalList;
import com.example.zoo.containers.impl.ThingList;
import com.example.zoo.service.ReportService;
import com.example.zoo.service.ZooService;
import com.example.zoo.service.dto.ZooSummaryDto;
import com.example.zoo.vet.VeterinaryClinic;
import org.junit.jupiter.api.*;
import org.springframework.beans.factory.NoSuchBeanDefinitionException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.context.ApplicationContext;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.argThat;
import static org.mockito.Mockito.*;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.PrintStream;
import java.util.List;

@SpringBootTest
class ZooApplicationTests {

    @Autowired
    private ZooService zooService;

    @Autowired
    private ReportService reportService;

    @Autowired
    private VeterinaryClinic clinic;

    @Autowired
    private ApplicationContext context;

    private PrintStream originalOut;
    private PrintStream originalErr;
    private java.io.InputStream originalIn;
    private ByteArrayOutputStream out;
    private ByteArrayOutputStream err;

    @BeforeEach
    void setupStreams() {
        originalOut = System.out;
        originalErr = System.err;
        originalIn = System.in;
        out = new ByteArrayOutputStream();
        err = new ByteArrayOutputStream();
        System.setOut(new PrintStream(out));
        System.setErr(new PrintStream(err));
    }

    @AfterEach
    void restoreStreams() {
        System.setOut(originalOut);
        System.setErr(originalErr);
        System.setIn(originalIn);
    }

    private void feedInput(String text) {
        System.setIn(new ByteArrayInputStream(text.getBytes()));
    }

    private String stdout() {
        return out.toString();
    }
    // ---------------------------------------------------

    @Test
    @DisplayName("Контекст Spring загружается и бины доступны")
    void contextLoads() {
        Assertions.assertNotNull(zooService);
        Assertions.assertNotNull(reportService);
        Assertions.assertNotNull(clinic);
    }

    @Test
    @DisplayName("Проверка работы контейнеров AnimalList и ThingList")
    void animalAndThingListsTest() {
        var animalList = new AnimalList();
        var thingList = new ThingList();

        var rabbit = new Rabbit("Бакс", 2, 10, 7);
        var computer = new Computer(3002, "ПК-учёт");

        animalList.add(rabbit);
        thingList.add(computer);

        assertEquals(1, animalList.findAll().size());
        assertEquals(1, thingList.findAll().size());

        // Проверка неизменяемости списка
        assertThrows(UnsupportedOperationException.class, () -> animalList.findAll().add(rabbit));
        assertThrows(UnsupportedOperationException.class, () -> thingList.findAll().clear());
    }

    @Test
    @DisplayName("Приём животных через ZooService и отчёт по ZooSummaryDto")
    void fullFlow_addAnimalsAndThings() {
        boolean m1 = zooService.admit(new Monkey("Чича", 3, 1001, 8)); // ok
        boolean r1 = zooService.admit(new Rabbit("Бакс", 1, 1002, 6)); // ok
        boolean t1 = zooService.admit(new Tiger("Шерхан", 7, 2001, 9)); // reject
        boolean w1 = zooService.admit(new Wolf("Серый", 5, 2002, 6));   // ok

        assertTrue(m1);
        assertTrue(r1);
        assertFalse(t1);
        assertTrue(w1);

        zooService.addThing(new Table(3001, "Стол"));
        zooService.addThing(new Computer(3002, "ПК"));

        ZooSummaryDto summary = reportService.buildSummary();
        assertEquals(3, summary.animalsCount());
        assertEquals(9, summary.totalFoodKgPerDay());
        assertTrue(summary.interactiveNames().contains("Monkey(Чича)"));
        assertEquals(5, summary.allInventory().size());
    }

    @Test
    @DisplayName("Отчёт без данных возвращает пустые коллекции")
    void reportWithEmptyData() {
        var emptyAnimalList = new AnimalList();
        var emptyThingList = new ThingList();
        var localReport = new ReportService(emptyAnimalList, emptyThingList);
        var result = localReport.buildSummary();

        assertNotNull(result);
        assertEquals(0, result.animalsCount());
        assertEquals(0, result.totalFoodKgPerDay());
        assertTrue(result.allInventory().isEmpty());
    }

    @Test
    @DisplayName("Проверка доменных классов животных и вещей")
    void domainClassesTest() {
        var monkey = new Monkey("Чича", 3, 1001, 8);
        var rabbit = new Rabbit("Бакс", 2, 1002, 7);
        var tiger = new Tiger("Шерхан", 7, 2001, 9);
        var wolf = new Wolf("Серый", 5, 2002, 6);
        var table = new Table(3003, "Стол");
        var computer = new Computer(3004, "ПК");

        assertEquals("Чича", monkey.getName());
        assertEquals(3, monkey.getFoodKgPerDay());
        assertTrue(rabbit.getDisplayName().contains("Rabbit"));
        assertTrue(tiger.getDisplayName().contains("Tiger"));
        assertEquals(5, wolf.getFoodKgPerDay());
        assertTrue(table.getDisplayName().contains("Table"));
        assertTrue(computer.getDisplayName().contains("Computer"));
    }

    @Test
    @DisplayName("Отрицательные значения доброты и опасности выбрасывают исключения")
    void invalidKindnessAndDangerThrows() {
        assertThrows(IllegalArgumentException.class, () -> new Rabbit("Bad", 1, 10, -1));
        assertThrows(IllegalArgumentException.class, () -> new Rabbit("Bad", 1, 10, 11));
        assertThrows(IllegalArgumentException.class, () -> new Tiger("BadT", 6, 20, -1));
        assertThrows(IllegalArgumentException.class, () -> new Tiger("BadT", 6, 20, 11));
    }

    @Test
    @DisplayName("Тест ветклиники: травоядные и хищники принимаются/отклоняются корректно")
    void veterinaryClinicPolicyTest() {
        var goodRabbit = new Rabbit("Добрый", 2, 101, 7);
        var badRabbit = new Rabbit("Злой", 2, 102, 1);
        var safeWolf = new Wolf("Спокойный", 5, 103, 6);
        var dangerousTiger = new Tiger("Агрессивный", 6, 104, 9);

        assertTrue(clinic.accept(goodRabbit));
        assertFalse(clinic.accept(badRabbit));
        assertTrue(clinic.accept(safeWolf));
        assertFalse(clinic.accept(dangerousTiger));
    }

    @Test
    @DisplayName("ZooService корректно работает при отклонении животных ветклиникой")
    void zooServiceRejectsAnimals() {
        var animalList = mock(AnimalList.class);
        var thingList = mock(ThingList.class);
        var clinicMock = mock(VeterinaryClinic.class);
        var service = new ZooService(animalList, thingList, clinicMock);

        var tiger = new Tiger("Шерхан", 7, 2001, 9);
        when(clinicMock.accept(tiger)).thenReturn(false);

        boolean accepted = service.admit(tiger);
        assertFalse(accepted);
        verify(animalList, never()).add(any());
    }

    @Test
    @DisplayName("ZooService добавляет вещи в контейнер")
    void zooServiceAddThingTest() {
        var animalList = mock(AnimalList.class);
        var thingList = mock(ThingList.class);
        var clinicMock = mock(VeterinaryClinic.class);
        var service = new ZooService(animalList, thingList, clinicMock);

        var table = new Table(1, "Test");
        service.addThing(table);
        verify(thingList, times(1)).add(table);
    }

    @Test
    @DisplayName("Проверка работы отчёта с минимальным порогом доброты")
    void reportKindnessThresholdTest() {
        var animals = new AnimalList();
        animals.add(new Rabbit("Середнячок", 2, 1001, 6));
        animals.add(new Monkey("Чича", 3, 1002, 8));
        var things = new ThingList();
        var report = new ReportService(animals, things);
        var s = report.buildSummary();

        assertEquals(2, s.animalsCount());
        assertTrue(s.interactiveNames().contains("Monkey(Чича)"));
    }

    @Test
    @DisplayName("Проверка работы clear в ThingList (пустой)")
    void clearOnEmptyThingList() {
        var list = new ThingList();
        list.clear();
        assertEquals(0, list.size());
    }

    @Test
    @DisplayName("Проверка работы clear в ThingList (с элементами)")
    void clearThingList() {
        var list = new ThingList();
        list.add(new Table(10, "T"));
        list.add(new Computer(11, "C"));
        assertEquals(2, list.size());

        list.clear();
        assertEquals(0, list.size());

        list.clear(); // повторный clear на пустом списке
        assertEquals(0, list.size());

        list.add(new Table(12, "T2"));
        assertEquals(1, list.size());
    }

    @Test
    @DisplayName("Проверка работы clear в AnimalList (пустой)")
    void clearOnEmptyAnimalList() {
        var list = new AnimalList();
        list.clear();
        assertEquals(0, list.size());
    }

    @Test
    @DisplayName("Проверка работы clear в AnimalList (с элементами)")
    void clearAnimalList() {
        var list = new AnimalList();
        list.add(new Rabbit("R", 1, 1, 7));
        list.add(new Tiger("T", 2, 2, 5));
        list.add(new Monkey("M", 1, 3, 8));
        assertEquals(3, list.size());

        list.clear();
        assertEquals(0, list.size());

        list.clear(); // повторный clear
        assertEquals(0, list.size());

        list.add(new Monkey("M2", 1, 4, 7));
        assertEquals(1, list.size());
    }

    @Test
    @DisplayName("Бин ConsoleRunner отсутствует по умолчанию (app.console.enabled=false в тестовой среде)")
    void consoleRunnerBeanAbsentByDefault() {
        assertThrows(NoSuchBeanDefinitionException.class,
                () -> context.getBean(ConsoleRunner.class));
    }

    @Test
    @DisplayName("ConsoleRunner: быстрый выход по 0")
    void consoleRunner_quickExit() throws Exception {
        // Моки сервисов
        var zoo = mock(ZooService.class);
        var report = mock(ReportService.class);
        var runner = new ConsoleRunner(zoo, report);

        feedInput("0\n");
        assertDoesNotThrow(() -> runner.run());
        String o = stdout();
        assertTrue(o.contains("Выберите действие"));
        assertTrue(o.contains("Программа завершена"));
    }

    @Test
    @DisplayName("ConsoleRunner: добавление Rabbit — принято ветклиникой")
    void consoleRunner_addRabbitAccepted() throws Exception {
        var zoo = mock(ZooService.class);
        var report = mock(ReportService.class);
        var runner = new ConsoleRunner(zoo, report);

        // 1 → Rabbit → Бакс → 2 → 101 → 7 → 0
        feedInput(String.join("\n", "1", "Rabbit", "Бакс", "2", "101", "7", "0") + "\n");

        when(zoo.admit(any(Rabbit.class))).thenReturn(true);

        runner.run();

        verify(zoo, times(1)).admit(any(Rabbit.class));
        assertTrue(stdout().contains("Животное принято в зоопарк"));
    }

    @Test
    @DisplayName("ConsoleRunner: добавление Tiger — отклонено ветклиникой")
    void consoleRunner_addTigerRejected() throws Exception {
        var zoo = mock(ZooService.class);
        var report = mock(ReportService.class);
        var runner = new ConsoleRunner(zoo, report);

        // 1 → Tiger → Шерхан → 7 → 2001 → 9 → 0
        feedInput(String.join("\n", "1", "Tiger", "Шерхан", "7", "2001", "9", "0") + "\n");

        when(zoo.admit(any(Tiger.class))).thenReturn(false);

        runner.run();

        verify(zoo, times(1)).admit(any(Tiger.class));
        assertTrue(stdout().contains("Ветеринар не одобрил животное"));
    }

    @Test
    @DisplayName("ConsoleRunner: неизвестный тип животного")
    void consoleRunner_addUnknownAnimalType() throws Exception {
        var zoo = mock(ZooService.class);
        var report = mock(ReportService.class);
        var runner = new ConsoleRunner(zoo, report);

        // 1 → Dragon → Имя → 1 → 100 → 5 → 0
        feedInput(String.join("\n", "1", "Dragon", "Имя", "1", "100", "5", "0") + "\n");

        runner.run();

        verify(zoo, never()).admit(any());
        assertTrue(stdout().contains("Неизвестный тип животного"));
    }

    @Test
    @DisplayName("ConsoleRunner: добавление вещи Table")
    void consoleRunner_addThingTable() throws Exception {
        var zoo = mock(ZooService.class);
        var report = mock(ReportService.class);
        var runner = new ConsoleRunner(zoo, report);

        // 2 → Table → Стол → 3001 → 0
        feedInput(String.join("\n", "2", "Table", "Стол", "3001", "0") + "\n");

        runner.run();

        verify(zoo, times(1)).addThing(argThat(t -> t instanceof Table && t.getNumber() == 3001));
        assertTrue(stdout().contains("Вещь добавлена в инвентарь"));
    }

    @Test
    @DisplayName("ConsoleRunner: добавление вещи Computer")
    void consoleRunner_addThingComputer() throws Exception {
        var zoo = mock(ZooService.class);
        var report = mock(ReportService.class);
        var runner = new ConsoleRunner(zoo, report);

        // 2 → Computer → ПК → 3002 → 0
        feedInput(String.join("\n", "2", "Computer", "ПК", "3002", "0") + "\n");

        runner.run();

        verify(zoo, times(1)).addThing(argThat(t -> t instanceof Computer && t.getNumber() == 3002));
        assertTrue(stdout().contains("Вещь добавлена в инвентарь"));
    }

    @Test
    @DisplayName("ConsoleRunner: неизвестный тип вещи")
    void consoleRunner_addUnknownThingType() throws Exception {
        var zoo = mock(ZooService.class);
        var report = mock(ReportService.class);
        var runner = new ConsoleRunner(zoo, report);

        // 2 → Sofa → Диван → 999 → 0
        feedInput(String.join("\n", "2", "Sofa", "Диван", "999", "0") + "\n");

        runner.run();

        verify(zoo, never()).addThing(any());
        assertTrue(stdout().contains("Неизвестный тип вещи"));
    }

    @Test
    @DisplayName("ConsoleRunner: неверный выбор меню")
    void consoleRunner_invalidMenuChoice() throws Exception {
        var zoo = mock(ZooService.class);
        var report = mock(ReportService.class);
        var runner = new ConsoleRunner(zoo, report);
        feedInput(String.join("\n", "x", "0") + "\n");

        runner.run();

        String o = stdout();
        assertTrue(o.contains("Неверный выбор, попробуйте снова"));
        assertTrue(o.contains("Программа завершена"));
    }
}
