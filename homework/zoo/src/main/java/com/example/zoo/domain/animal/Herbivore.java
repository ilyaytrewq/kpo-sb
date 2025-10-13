package com.example.zoo.domain.animal;

public abstract class Herbivore extends Animal {
    private final int kindness;


    protected Herbivore(String name, int foodKgPerDay, int inventoryNumber, int kindness) {
        super(name, foodKgPerDay, inventoryNumber);
        if (kindness < 0 || kindness > 10) throw new IllegalArgumentException("kindness must be 0..10");
        this.kindness = kindness;
    }


    public int getKindness() { return kindness; }
}
