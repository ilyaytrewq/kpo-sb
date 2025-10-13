package com.example.zoo.domain.animal;

public abstract class Predator extends Animal {
    private final int danger;
    protected Predator(String name, int foodKgPerDay, int inventoryNumber, int danger) {
        super(name, foodKgPerDay, inventoryNumber);
        if (danger < 0 || danger > 10) throw new IllegalArgumentException("danger must be 0..10");
        this.danger = danger;
    }


    public int getDanger() { return danger; }
}