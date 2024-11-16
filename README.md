# Yahfaz LINE Bot

Yahfaz is a bot that supports you in memorizing the Quran using a spaced repetition system, helping you stay consistent even with a busy schedule.

## Main Features

### Learn

Use the Learn feature to log Quran pages you've memorized. Yahfaz accepts entries one page at a time.

### Review

Yahfaz will schedule reviews for you based on spaced repetition principles, reminding you when it's time to revisit a page. To see which pages are scheduled for review, use the Show command. When you're ready, use the Review command and assess your memorization for each page:

- Easy: 0-1 mistakes (review interval increases).
- Good: 2-3 mistakes.
- Hard: 4+ mistakes (review interval shortens).

With Yahfaz, you can keep track of your progress and review efficiently, ensuring long-term retention.

<!-- GETTING STARTED -->
## Getting Started

To run this project locally, you need several things to do.

### Prerequisites

Install preqrequisites app that needed to install this project.
* [MySQL][MySQL-url]
* [Golang][Go-url]
* [Ngrok][Ngrok-url]

### Installation

1. Clone the repo
   ```sh
   git clone https://github.com/AkmalArifin/srq-line-bot.git
   ```
2. Create new database in MySQL and import from `db.sql`
3. Create .env file
4. Run the server in backend folder
   ```sh
   go run .
   ```
5. Run ngrok for the port
   ```sh
   ngrok http 8080
   ```
6. Copy the ngrok url to LINE webhook url and add `callback` endpoint


<!-- MARKDOWN LIST & IMAGES -->
[Golang]: https://img.shields.io/badge/Go-00ADD8?logo=Go&logoColor=white&style=for-the-badge
[Go-url]: https://go.dev/
[MySQL]: https://img.shields.io/badge/MySQL-4479A1?style=for-the-badge&logo=mysql&logoColor=white
[MySQL-url]: https://www.mysql.com/
[Nodejs-url]: https://nodejs.org/en
[Ngrok-url]: https://ngrok.com/
