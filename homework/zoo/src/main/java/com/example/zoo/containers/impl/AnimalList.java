package com.example.zoo.containers.impl;

import com.example.zoo.containers.IAnimalList;
import com.example.zoo.domain.animal.Animal;
import org.springframework.stereotype.Repository;
import java.util.*;

@Repository
public class AnimalList implements IAnimalList {
    private final List<Animal> data = new ArrayList<>();
    @Override public void add(Animal animal) { data.add(animal); }
    @Override public List<Animal> findAll() { return Collections.unmodifiableList(data); }

    @Override public void clear() {
        data.clear();
    }
    @Override public int size() { return data.size(); }
}