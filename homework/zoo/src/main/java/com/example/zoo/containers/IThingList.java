package com.example.zoo.containers;

import com.example.zoo.domain.thing.Thing;
import java.util.List;

public interface IThingList {
    void add(Thing thing);

    List<Thing> findAll();

    void clear();

    int size();
}
