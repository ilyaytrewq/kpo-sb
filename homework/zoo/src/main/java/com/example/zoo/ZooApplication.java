package com.example.zoo;


import com.example.zoo.domain.animal.*;
import com.example.zoo.domain.thing.*;
import com.example.zoo.service.ReportService;
import com.example.zoo.service.ZooService;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;


@SpringBootApplication
public class ZooApplication {
    public static void main(String[] args) { SpringApplication.run(ZooApplication.class, args); }
}