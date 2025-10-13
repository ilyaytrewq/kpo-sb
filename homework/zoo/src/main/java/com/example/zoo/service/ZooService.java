package com.example.zoo.service;


import com.example.zoo.domain.animal.Animal;
import com.example.zoo.domain.thing.Thing;
import com.example.zoo.containers.impl.AnimalList;
import com.example.zoo.containers.impl.ThingList;
import com.example.zoo.vet.VeterinaryClinic;
import org.springframework.stereotype.Service;


@Service
public class ZooService {
    private final AnimalList animalRepo;
    private final ThingList thingRepo;
    private final VeterinaryClinic clinic;


    public ZooService(AnimalList animalRepo, ThingList thingRepo, VeterinaryClinic clinic) {
        this.animalRepo = animalRepo; this.thingRepo = thingRepo; this.clinic = clinic;
    }

    public boolean admit(Animal animal) {
        if (clinic.accept(animal)) { animalRepo.add(animal); return true; }
        return false;
    }


    public void addThing(Thing thing) { thingRepo.add(thing); }
}