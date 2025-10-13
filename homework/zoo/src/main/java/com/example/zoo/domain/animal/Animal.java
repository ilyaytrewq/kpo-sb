package com.example.zoo.domain.animal;

import com.example.zoo.domain.IAlive;
import com.example.zoo.domain.IInventory;


public abstract class Animal implements IAlive, IInventory {
    private final String name;
    private final int foodKgPerDay;
    private final int inventoryNumber;


    protected Animal(String name, int foodKgPerDay, int inventoryNumber) {
        this.name = name;
        this.foodKgPerDay = foodKgPerDay;
        this.inventoryNumber = inventoryNumber;
    }


    public String getName() { return name; }
    @Override public int getFoodKgPerDay() { return foodKgPerDay; }
    @Override public int getNumber() { return inventoryNumber; }
    @Override public String getDisplayName() { return getClass().getSimpleName() + "(" + name + ")"; }
}