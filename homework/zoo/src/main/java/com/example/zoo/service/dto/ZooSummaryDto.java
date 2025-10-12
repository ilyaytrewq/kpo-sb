package com.example.zoo.service.dto;


import java.util.List;


public record ZooSummaryDto(int animalsCount, int totalFoodKgPerDay, List<String> interactiveNames,
                            List<InventoryDto> allInventory) {}