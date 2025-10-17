package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// ===== 1. Загрузка переменных окружения =====
	_ = godotenv.Load()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// fallback — для локальной отладки (замени пароль на свой)
		dsn = "postgres://postgres:YOUR_PASSWORD@localhost:5432/todo?sslmode=disable&connect_timeout=5"
	}

	// ===== 2. Подключение к БД =====
	db, err := openDB(dsn)
	if err != nil {
		log.Fatalf("openDB error: %v", err)
	}
	defer db.Close()

	repo := NewRepo(db)

	// ===== 3. Базовые вставки =====
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	titles := []string{"Сделать ПЗ №5", "Купить кофе", "Проверить отчёты"}
	for _, title := range titles {
		id, err := repo.CreateTask(ctx, title)
		if err != nil {
			log.Fatalf("CreateTask error: %v", err)
		}
		log.Printf("Inserted task id=%d (%s)", id, title)
	}

	// ===== 4. Список задач =====
	ctxList, cancelList := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelList()
	tasks, err := repo.ListTasks(ctxList)
	if err != nil {
		log.Fatalf("ListTasks error: %v", err)
	}
	fmt.Println("=== Tasks ===")
	for _, t := range tasks {
		fmt.Printf("#%d | %-24s | done=%-5v | %s\n",
			t.ID, t.Title, t.Done, t.CreatedAt.Format(time.RFC3339))
	}

	// ===== 5. ДОПОЛНИТЕЛЬНЫЕ ЗАДАНИЯ =====
	{
		// 5.1 Обновим задачу id=1 как выполненную
		ctxUpd, cancelUpd := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelUpd()
		if _, err := db.ExecContext(ctxUpd, `UPDATE tasks SET done = TRUE WHERE id = $1`, 1); err != nil {
			log.Printf("mark done error: %v", err)
		}

		// 5.2 Найдём задачу по ID
		ctxFind, cancelFind := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelFind()
		if task, err := repo.FindByID(ctxFind, 1); err == nil && task != nil {
			fmt.Printf("FindByID(1): #%d | %s | done=%v\n", task.ID, task.Title, task.Done)
		} else if err != nil {
			log.Printf("FindByID error: %v", err)
		} else {
			fmt.Println("FindByID(1): not found")
		}

		// 5.3 Список по статусу
		ctxDone, cancelDone := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelDone()
		undone, err := repo.ListDone(ctxDone, false)
		if err != nil {
			log.Printf("ListDone(false) error: %v", err)
		}
		doneList, err := repo.ListDone(ctxDone, true)
		if err != nil {
			log.Printf("ListDone(true) error: %v", err)
		}
		fmt.Printf("ListDone(false): %d задач(и)\n", len(undone))
		fmt.Printf("ListDone(true):  %d задач(и)\n", len(doneList))

		// 5.4 Массовая вставка
		ctxMany, cancelMany := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancelMany()
		if err := repo.CreateMany(ctxMany, []string{"Ещё одна", "И ещё одна"}); err != nil {
			log.Printf("CreateMany error: %v", err)
		} else {
			fmt.Println("CreateMany: вставили 2 задачи")
		}

		// 5.5 Финальный список
		ctxList2, cancelList2 := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancelList2()
		tasks2, err := repo.ListTasks(ctxList2)
		if err != nil {
			log.Fatalf("ListTasks(2) error: %v", err)
		}
		fmt.Println("=== Tasks (after extras) ===")
		for _, t := range tasks2 {
			fmt.Printf("#%d | %-24s | done=%-5v | %s\n",
				t.ID, t.Title, t.Done, t.CreatedAt.Format(time.RFC3339))
		}
	}
}
