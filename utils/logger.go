package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局日志变量，以后在任何文件里都可以直接用 utils.Logger.Info() 记录日志
var Logger *zap.Logger

func InitLogger() {
	// 1. 配置 Lumberjack 实现日志自动切割 (伐木工)
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "logs/app.log", // 日志文件的位置
		MaxSize:    10,             // 每个日志文件最大 10 MB，超过就自动切割
		MaxBackups: 30,             // 最多保留 30 个备份文件
		MaxAge:     30,             // 最多保留 30 天的日志
		Compress:   true,           // 是否压缩旧日志 (节省服务器磁盘空间)
	}
	writeSyncer := zapcore.AddSync(lumberJackLogger)

	// 2. 配置日志输出格式 (JSON 格式，方便日后接入 ELK 日志分析系统)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder   // 时间格式：2024-04-12T10:00:00.000+0800
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 级别格式：大写的 INFO, ERROR
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 3. 同时输出到控制台 (方便你在本地开发时看黑框框)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleSyncer := zapcore.AddSync(os.Stdout)

	// 4. 将输出到文件和输出到控制台合并 (Tee 模式)
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel),           // 写入文件的策略：INFO 及以上级别
		zapcore.NewCore(consoleEncoder, consoleSyncer, zapcore.DebugLevel), // 控制台输出：DEBUG 及以上级别
	)

	// 5. 生成最终的 Logger，并开启调用者信息记录 (AddCaller 能帮你记录是哪行代码打的日志)
	Logger = zap.New(core, zap.AddCaller())

	// 替换全局的 zap 实例
	zap.ReplaceGlobals(Logger)

	Logger.Info("✅ 企业级 Zap 日志模块初始化成功！")
}
