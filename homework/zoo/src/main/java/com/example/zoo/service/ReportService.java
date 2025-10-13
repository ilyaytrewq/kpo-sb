package com.example.zoo.service;


import com.example.zoo.domain.IInventory;
import com.example.zoo.domain.animal.Animal;
import com.example.zoo.domain.animal.Herbivore;
import com.example.zoo.containers.impl.AnimalList;
import com.example.zoo.containers.impl.ThingList;
import com.example.zoo.service.dto.ZooSummaryDto;
import com.example.zoo.service.dto.InventoryDto;
import org.springframework.stereotype.Service;


import java.util.ArrayList;
import java.util.List;


@Service
public class ReportService {
    private final AnimalList animalRepo;
    private final ThingList thingRepo;
    private final int kindnessThreshold;


    public ReportService(AnimalList animalRepo, ThingList thingRepo) {
        this.animalRepo = animalRepo; this.thingRepo = thingRepo; this.kindnessThreshold = 5;
    }

    public ZooSummaryDto buildSummary() {
        var animals = animalRepo.findAll();
        int totalFood = animals.stream().mapToInt(Animal::getFoodKgPerDay).sum();
        List<String> interactive = animals.stream()
                .filter(a -> a instanceof Herbivore h && h.getKindness() >= kindnessThreshold)
                .map(IInventory::getDisplayName)
                .toList();

        List<IInventory> all = new ArrayList<>();
        all.addAll(animals);
        all.addAll(thingRepo.findAll());

        List<InventoryDto> items = all.stream()
                .map(x -> new InventoryDto(x.getNumber(), x.getDisplayName()))
                .toList();

        return new ZooSummaryDto(animals.size(), totalFood, interactive, items);
    }

}