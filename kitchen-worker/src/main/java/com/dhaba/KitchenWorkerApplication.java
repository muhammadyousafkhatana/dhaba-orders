package com.dhaba;

import java.sql.DriverManager;
import java.util.Map;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.boot.CommandLineRunner;
import org.springframework.context.annotation.Bean;

@SpringBootApplication
public class KitchenWorkerApplication {
    public static void main(String[] args) {
        SpringApplication.run(KitchenWorkerApplication.class, args);
    }

    @Bean
    CommandLineRunner kitchen(StringRedisTemplate redis) {
        return args -> {
            String url = "jdbc:postgresql://" +
                System.getenv().getOrDefault("DB_HOST", "postgres") + ":5432/dhaba";
            String user = System.getenv().getOrDefault("DB_USER", "dhaba");
            String password = System.getenv().getOrDefault("DB_PASSWORD", "dhaba-dev-change-this");

            while (true) {
                String orderId = redis.opsForList().rightPop("kitchen:orders",
                    java.time.Duration.ofSeconds(5));
                if (orderId == null) continue;

                try (var conn = DriverManager.getConnection(url, user, password);
                     var statement = conn.prepareStatement(
                         "UPDATE orders SET status='COOKING', updated_at=NOW() WHERE id=? AND status='NEW'")) {
                    statement.setLong(1, Long.parseLong(orderId));
                    int updated = statement.executeUpdate();
                    if (updated == 1) {
                        Thread.sleep(5000);
                        try (var ready = conn.prepareStatement(
                            "UPDATE orders SET status='READY', updated_at=NOW() WHERE id=? AND status='COOKING'")) {
                            ready.setLong(1, Long.parseLong(orderId));
                            ready.executeUpdate();
                        }
                        System.out.println("Order " + orderId + " is READY");
                    }
                } catch (Exception e) {
                    System.err.println("Kitchen processing failed: " + e.getMessage());
                    redis.opsForList().leftPush("kitchen:orders", orderId);
                    Thread.sleep(3000);
                }
            }
        };
    }
}
