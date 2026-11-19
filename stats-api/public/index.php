<?php

header('Content-Type: application/json');

$host = getenv('DB_HOST') ?: 'postgres';
$db = getenv('DB_NAME') ?: 'dhaba';
$user = getenv('DB_USER') ?: 'dhaba';
$password = getenv('DB_PASSWORD') ?: 'dhaba-dev-change-this';

try {
    $pdo = new PDO(
        "pgsql:host=$host;port=5432;dbname=$db",
        $user,
        $password,
        [PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION]
    );

    $path = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);

    if ($path === '/health') {
        echo json_encode(['status' => 'healthy']);
        exit;
    }

    if ($path === '/stats') {
        $summary = $pdo->query("
            SELECT COUNT(*) AS total_orders,
                   COUNT(*) FILTER (WHERE status='NEW') AS new_orders,
                   COUNT(*) FILTER (WHERE status='COOKING') AS cooking_orders,
                   COUNT(*) FILTER (WHERE status='READY') AS ready_orders,
                   COALESCE(SUM(total_amount),0) AS revenue
            FROM orders
        ")->fetch(PDO::FETCH_ASSOC);

        echo json_encode($summary);
        exit;
    }

    http_response_code(404);
    echo json_encode(['error' => 'Not found']);
} catch (Throwable $e) {
    http_response_code(503);
    echo json_encode(['error' => 'Database unavailable']);
}
