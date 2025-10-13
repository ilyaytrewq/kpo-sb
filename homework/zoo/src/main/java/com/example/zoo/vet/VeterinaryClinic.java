package com.example.zoo.vet;


import com.example.zoo.domain.animal.Animal;
import com.example.zoo.domain.animal.Herbivore;
import com.example.zoo.domain.animal.Predator;
import org.springframework.stereotype.Component;
import java.util.Random;

@Component
public class VeterinaryClinic implements IVeterinaryClinic {
    private static final Random RANDOM = new Random();

    @Override
    public boolean accept(Animal animal) {
        if (animal instanceof Herbivore h) {
            return h.getKindness() >= 2;
        } else if (animal instanceof Predator p) {
            return p.getDanger() <= 7;
        }
        return false;
    }


}