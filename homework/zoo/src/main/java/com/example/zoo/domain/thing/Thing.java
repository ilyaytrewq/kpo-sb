package com.example.zoo.domain.thing;

import com.example.zoo.domain.IInventory;

public abstract class Thing implements IInventory {
    private final int number;
    private final String name;

    protected Thing(int number, String name) {
        this.number = number; this.name = name;
    }

    @Override public int getNumber() { return number; }
    public String getName() { return name; }
    @Override public String getDisplayName() { return getClass().getSimpleName() + "(" + name + ")"; }
}