package com.example.zoo.vet;
import com.example.zoo.domain.animal.Animal;

public interface IVeterinaryClinic {
    boolean accept(Animal animal);
}