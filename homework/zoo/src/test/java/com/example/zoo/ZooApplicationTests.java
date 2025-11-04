// java
package com.example.zoo;

import com.example.zoo.containers.impl.AnimalList;
import com.example.zoo.containers.impl.ThingList;
import com.example.zoo.domain.animal.Monkey;
import com.example.zoo.domain.animal.Rabbit;
import com.example.zoo.domain.animal.Tiger;
import com.example.zoo.domain.animal.Wolf;
import com.example.zoo.domain.thing.Computer;
import com.example.zoo.domain.thing.Table;
import com.example.zoo.service.ReportService;
import com.example.zoo.service.ZooService;
import com.example.zoo.vet.VeterinaryClinic;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@SpringBootTest
class ZooApplicationTests {

    @Autowired
    private ZooService zooService;

    @Autowired
    private ReportService reportService;

    @Autowired
    private VeterinaryClinic clinic;

    @Test
    @DisplayName("Контекст загружается, сервисы и клиника доступны")
    void contextLoads() {
        assertNotNull(zooService);
        assertNotNull(reportService);
        assertNotNull(clinic);
    }

    @Test
    @DisplayName("AnimalList: add/findAll/clear/size и неизменяемая выдача")
    void animalListBasics() {
        var list = new AnimalList();
        var r = new Rabbit("Бакс", 2, 10, 7);
        list.add(r);

        assertEquals(1, list.size());
        assertEquals(1, list.findAll().size());
        assertThrows(UnsupportedOperationException.class, () -> list.findAll().add(r));

        list.clear();
        assertEquals(0, list.size());
        assertTrue(list.findAll().isEmpty());
    }

    @Test
    @DisplayName("ThingList: add/findAll/clear/size и неизменяемая выдача")
    void thingListBasics() {
        var list = new ThingList();
        var t = new Computer(100, "ПК");

        list.add(t);
        assertEquals(1, list.size());
        assertEquals(1, list.findAll().size());
        assertThrows(UnsupportedOperationException.class, () -> list.findAll().clear());

        list.clear();
        assertEquals(0, list.size());
        assertTrue(list.findAll().isEmpty());
    }

    @Test
    @DisplayName("Доменные классы: геттеры и displayName")
    void domainClasses() {
        var monkey = new Monkey("Чича", 3, 1001, 8);
        var rabbit = new Rabbit("Бакс", 2, 1002, 7);
        var tiger  = new Tiger("Шерхан", 7, 2001, 9);
        var wolf   = new Wolf("Серый", 5, 2002, 6);
        var table  = new Table(3003, "Стол");
        var pc     = new Computer(3004, "ПК");

        assertEquals("Чича", monkey.getName());
        assertEquals(3, monkey.getFoodKgPerDay());
        assertTrue(rabbit.getDisplayName().contains("Rabbit"));
        assertTrue(tiger.getDisplayName().contains("Tiger"));
        assertEquals(5, wolf.getFoodKgPerDay());
        assertTrue(table.getDisplayName().contains("Table"));
        assertTrue(pc.getDisplayName().contains("Computer"));
    }

    @Test
    @DisplayName("Валидация: доброта/опасность вне [0..10] — исключение")
    void validationRanges() {
        assertThrows(IllegalArgumentException.class, () -> new Rabbit("Bad", 1, 10, -1));
        assertThrows(IllegalArgumentException.class, () -> new Rabbit("Bad", 1, 10, 11));
        assertThrows(IllegalArgumentException.class, () -> new Tiger("BadT", 6, 20, -1));
        assertThrows(IllegalArgumentException.class, () -> new Tiger("BadT", 6, 20, 11));
    }

    @Test
    @DisplayName("Ветклиника: правила принятия/отклонения")
    void veterinaryPolicy() {
        var okRabbit  = new Rabbit("Добряк", 2, 1, 7);
        var badRabbit = new Rabbit("Злюка", 2, 2, 1);
        var okWolf    = new Wolf("Спокойный", 5, 3, 6);
        var badTiger  = new Tiger("Агрессивный", 7, 4, 9);

        assertTrue(clinic.accept(okRabbit));
        assertFalse(clinic.accept(badRabbit));
        assertTrue(clinic.accept(okWolf));
        assertFalse(clinic.accept(badTiger));
    }

    @Test
    @DisplayName("ZooService: приём животных и добавление вещей")
    void zooServiceFlow() {
        assertTrue(zooService.admit(new Monkey("Чича", 3, 1001, 8)));
        assertTrue(zooService.admit(new Rabbit("Бакс", 1, 1002, 6)));
        assertFalse(zooService.admit(new Tiger("Шерхан", 7, 2001, 9)));
        assertTrue(zooService.admit(new Wolf("Серый", 5, 2002, 6)));

        zooService.addThing(new Table(3001, "Стол"));
        zooService.addThing(new Computer(3002, "ПК"));

        var s = reportService.buildSummary();
        assertEquals(3, s.animalsCount());
        assertEquals(9, s.totalFoodKgPerDay());
        assertTrue(s.interactiveNames().contains("Monkey(Чича)"));
        assertTrue(s.allInventory().stream().anyMatch(r -> r.number() == 3001 && r.name().contains("Table")));
        assertTrue(s.allInventory().stream().anyMatch(r -> r.number() == 3002 && r.name().contains("Computer")));
    }

    @Test
    @DisplayName("ZooService: отклонённое животное не добавляется")
    void rejectAnimalDoesNotAdd() {
        var animals = mock(AnimalList.class);
        var things  = mock(ThingList.class);
        var vet     = mock(VeterinaryClinic.class);
        var service = new ZooService(animals, things, vet);

        var tiger = new Tiger("Шерхан", 7, 9, 9);
        when(vet.accept(tiger)).thenReturn(false);

        assertFalse(service.admit(tiger));
        verify(animals, never()).add(any());
    }

    @Test
    @DisplayName("ZooService: addThing делегирует контейнеру")
    void addThingDelegatesToContainer() {
        var animals = mock(AnimalList.class);
        var things  = mock(ThingList.class);
        var vet     = mock(VeterinaryClinic.class);
        var service = new ZooService(animals, things, vet);

        var table = new Table(1, "T");
        service.addThing(table);
        verify(things, times(1)).add(table);
    }

    @Test
    @DisplayName("ReportService: пустые контейнеры → пустой отчёт")
    void reportEmpty() {
        var a = new AnimalList();
        var t = new ThingList();
        var report = new ReportService(a, t);
        var s = report.buildSummary();

        assertEquals(0, s.animalsCount());
        assertEquals(0, s.totalFoodKgPerDay());
        assertTrue(s.interactiveNames().isEmpty());
        assertTrue(s.allInventory().isEmpty());
    }
}