package com.example.zoo.containers.impl;


import com.example.zoo.containers.IThingList;
import com.example.zoo.domain.thing.Thing;
import org.springframework.stereotype.Repository;
import java.util.*;

@Repository
public class ThingList implements IThingList {
    private final List<Thing> data = new ArrayList<>();
    @Override public void add(Thing thing) { data.add(thing); }
    @Override public List<Thing> findAll() { return Collections.unmodifiableList(data); }
    @Override public void clear() {
        data.clear();
    }
    @Override public int size() { return data.size(); }
}