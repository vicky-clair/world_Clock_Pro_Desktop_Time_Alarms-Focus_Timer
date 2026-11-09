# Themida 保护与加壳配置指南

本项目支持使用 **Themida**（Oreans Technologies 出品的高级软件加固与代码虚拟化保护系统）对编译生成的 Windows PE 可执行文件（`.exe`）进行加密保护与反逆向加固。

---

## ⚠️ 重要安全与杀毒软件预警（必读）

1. **Themida 是“PE 加壳保护器”，不是“安装包制作工具”**：
   - Themida 的作用是保护 `.exe` 和 `.dll`，防止被逆向工程、内存篡改或调试。
   - **Themida 无法生成 `.msi` 安装包**（Windows Installer 安装包由 WiX Toolset 编译）。
   - **Themida 无法生成苹果 macOS / iOS 格式**（苹果平台采用 Mach-O 二进制和 `.app` / `.pkg` 打包，Themida 仅支持 Windows PE 格式）。
2. **极高概率触发杀毒软件启发式误报**：
   - Themida 注入了反调试、API 混淆与虚拟机代码段，这些特征经常被恶意软件加壳利用。因此，被 Themida 加壳后的程序，在卡巴斯基、Windows Defender、360、火绒等杀毒引擎下会被报为 `Heur.Packer.Themida`、`Malware-Cryptor.Themida` 或 `VHO:HackTool.Win64.Convagent.gen`。
   - 若要对外发布，**建议优先使用本项目原生提供的 Authenticode 数字签名打包方式（`build_windows.ps1`）**，这样报毒率为 0；若确实需要使用 Themida 加壳，请为加壳后的可执行文件购买正规机构颁发的 EV 代码签名证书（EV Code Signing Certificate）。
3. **针对 Go 语言特性的防崩溃配置建议**：
   - Go 语言自带复杂的底层运行时（Runtime 调度器、Goroutine 栈管理、垃圾回收器 GC 和信号处理机制）。
   - **请勿勾选**：“Anti-Debugger - Kernel Mode Check” 或对入口点代码进行过深的代码虚拟化（Code Virtualization），否则 Go 运行时初始化时可能发生 `Access Violation`（0xC0000005）异常退出。
   - **推荐勾选**：“Anti-Dump”、“Memory Protection - Basic” 以及针对主业务模块的混淆。

---

## 🛠️ 实操步骤

### 方式 1：一键脚本引导
直接在 PowerShell 中运行脚本：
```powershell
powershell -ExecutionPolicy Bypass -File .\packaging\themida\protect_themida.ps1
```
脚本会自动拉起 `Themida.exe` 并载入 `dist\小瑞综合时钟桌面.exe`。

### 方式 2：Themida 图形界面手动配置
1. 打开 `C:\BLWJ\Themida_3.2.6.0_nCh2CZZRNLW\Themida.exe`；
2. **Input Filename**：选择 `dist\小瑞综合时钟桌面.exe`；
3. **Output Filename**：设置为 `dist\小瑞综合时钟桌面_Protected.exe`；
4. **Protection Options**：
   - 勾选 `Compression`（压缩外壳，建议 Level: Normal）；
   - 勾选 `Anti-Dump Protection`；
   - 保持 `Code Virtualization` 为适度或仅针对特定函数；
5. 点击左侧工具栏的 **“Protect”** 按钮完成加壳；
6. 加壳完成后，重新对生成的 EXE 进行签名：
   ```powershell
   $cert = Get-ChildItem Cert:\CurrentUser\My -CodeSigningCert | Select-Object -First 1
   Set-AuthenticodeSignature -FilePath "dist\小瑞综合时钟桌面_Protected.exe" -Certificate $cert -HashAlgorithm SHA256
   ```
