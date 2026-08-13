package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

// ============================================================
// ۱. تنظیمات اولیه
// ============================================================

var (
	githubToken   = os.Getenv("GITHUB_TOKEN")
	repoOwner     = "beaconchain-horizon"
	repoName      = "horizon-core-engine"
	branch        = "main"
	localPath     = "./horizon-build"
	services      = []string{"core", "license", "switch"}
	servicePorts  = map[string]string{"core": "3000", "license": "4000", "switch": "8080"}
)

// ============================================================
// ۲. اتصال به گیت‌هاب
// ============================================================

func getGitHubClient() *github.Client {
	if githubToken == "" {
		log.Fatal("❌ GITHUB_TOKEN environment variable not set")
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// ============================================================
// ۳. دریافت آخرین کامیت
// ============================================================

func getLatestCommit(client *github.Client) (string, error) {
	ctx := context.Background()
	commits, _, err := client.Repositories.ListCommits(ctx, repoOwner, repoName, &github.CommitsListOptions{
		SHA: branch,
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		return "", err
	}
	if len(commits) == 0 {
		return "", fmt.Errorf("no commits found")
	}
	return *commits[0].SHA, nil
}

// ============================================================
// ۴. کلون کردن مخزن (با استفاده از GitHub API)
// ============================================================

func cloneRepo(client *github.Client) error {
	log.Println("📦 Cloning repository...")
	ctx := context.Background()

	// دریافت بایگانی مخزن به‌صورت ZIP
	archiveURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/zipball/%s", repoOwner, repoName, branch)
	req, err := http.NewRequestWithContext(ctx, "GET", archiveURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+githubToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download repo: %s", resp.Status)
	}

	// حذف پوشه قبلی
	os.RemoveAll(localPath)
	os.MkdirAll(localPath, 0755)

	// ذخیره فایل ZIP موقت
	zipPath := filepath.Join(localPath, "repo.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	io.Copy(zipFile, resp.Body)

	log.Println("✅ Repository downloaded successfully")
	return nil
}

// ============================================================
// ۵. کامپایل پروژه
// ============================================================

func buildServices() error {
	log.Println("🔨 Building services...")

	for _, svc := range services {
		log.Printf("  Building %s...", svc)
		cmd := exec.Command("go", "build", "-o", filepath.Join(localPath, svc, "horizon-"+svc), "./cmd/"+svc)
		cmd.Dir = localPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error building %s: %v", svc, err)
		}
	}
	log.Println("✅ All services built successfully")
	return nil
}

// ============================================================
// ۶. اجرای سرویس‌ها
// ============================================================

func runServices() {
	log.Println("🚀 Starting services...")

	for _, svc := range services {
		go func(s string) {
			binPath := filepath.Join(localPath, s, "horizon-"+s)
			cmd := exec.Command(binPath)
			cmd.Dir = filepath.Join(localPath, s)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			log.Printf("✅ %s started on port %s", s, servicePorts[s])
			if err := cmd.Run(); err != nil {
				log.Printf("❌ %s exited with error: %v", s, err)
			}
		}(svc)
	}
}

// ============================================================
// ۷. نمایش وضعیت سرویس‌ها و لینک‌ها
// ============================================================

func showDashboard() {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("  🌍 Horizon Ecosystem - All Services Running")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()
	fmt.Println("  📊 Dashboards:")
	fmt.Println("  ----------------------------------------")
	fmt.Println("  🔹 Bank Saderat Dashboard:")
	fmt.Println("     http://localhost:" + servicePorts["switch"] + "/bank-saderat-dashboard.html")
	fmt.Println()
	fmt.Println("  🔹 License Manager:")
	fmt.Println("     http://localhost:" + servicePorts["switch"] + "/license-manager.html")
	fmt.Println()
	fmt.Println("  🔹 Personal Panel:")
	fmt.Println("     http://localhost:" + servicePorts["switch"] + "/dashboard-personal.html")
	fmt.Println("  ----------------------------------------")
	fmt.Println()
	fmt.Println("  🔴 Press Ctrl+C to stop all services")
	fmt.Println()
}

// ============================================================
// ۸. تابع اصلی (Orchestrator)
// ============================================================

func main() {
	log.Println("========================================")
	log.Println("  🌍 Horizon Orchestrator v1.0")
	log.Println("  Automated Workflow & Deployment")
	log.Println("========================================")

	// ۱. اتصال به گیت‌هاب
	client := getGitHubClient()

	// ۲. دریافت آخرین کامیت
	commitSHA, err := getLatestCommit(client)
	if err != nil {
		log.Fatalf("❌ Failed to get latest commit: %v", err)
	}
	log.Printf("🔍 Latest commit: %s", commitSHA[:8])

	// ۳. کلون کردن مخزن
	if err := cloneRepo(client); err != nil {
		log.Fatalf("❌ Failed to clone repo: %v", err)
	}

	// ۴. کامپایل سرویس‌ها
	if err := buildServices(); err != nil {
		log.Fatalf("❌ Build failed: %v", err)
	}

	// ۵. اجرای سرویس‌ها
	runServices()

	// ۶. نمایش داشبورد
	time.Sleep(2 * time.Second) // منتظر می‌مانیم تا سرویس‌ها بالا بیایند
	showDashboard()

	// ۷. منتظر ورود کاربر برای خروج
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// ============================================================
// ۹. یادداشت: فایل go.mod (نیازمند این وابستگی‌ها)
// ============================================================
// go mod init horizon-orchestrator
// go get github.com/google/go-github/v68/github
// go get golang.org/x/oauth2
