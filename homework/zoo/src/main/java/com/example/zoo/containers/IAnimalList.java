package com.example.zoo.containers;


import java.util.List;
import com.example.zoo.domain.animal.Animal;


public interface IAnimalList {
    void add(Animal animal);
    List<Animal> findAll();
    void clear();

    int size();
}