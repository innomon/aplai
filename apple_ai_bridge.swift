import Foundation
import FoundationModels

public class AppleAISession {
    let session: LanguageModelSession
    
    init(instructions: String?) {
        let model = SystemLanguageModel.default
        self.session = LanguageModelSession(model: model, instructions: instructions)
    }
}

@_cdecl("CreateAppleAISession")
public func CreateAppleAISession(instructions: UnsafePointer<CChar>?) -> UnsafeMutableRawPointer? {
    let instStr = instructions != nil ? String(cString: instructions!) : nil
    let session = AppleAISession(instructions: instStr)
    return Unmanaged.passRetained(session).toOpaque()
}

@_cdecl("DestroyAppleAISession")
public func DestroyAppleAISession(sessionPtr: UnsafeMutableRawPointer) {
    let _: AppleAISession = Unmanaged.fromOpaque(sessionPtr).takeRetainedValue()
}

@_cdecl("AppleAISessionRespond")
public func AppleAISessionRespond(sessionPtr: UnsafeMutableRawPointer, prompt: UnsafePointer<CChar>) -> UnsafeMutablePointer<CChar> {
    let session = Unmanaged<AppleAISession>.fromOpaque(sessionPtr).takeUnretainedValue()
    let promptStr = String(cString: prompt)
    
    let semaphore = DispatchSemaphore(value: 0)
    var resultStr = ""
    
    Task {
        do {
            let response = try await session.session.respond(to: promptStr)
            resultStr = response.content
        } catch {
            resultStr = "ERROR: \(error.localizedDescription)"
        }
        semaphore.signal()
    }
    
    semaphore.wait()
    return strdup(resultStr)
}

public typealias AppleAITokenCallback = @convention(c) (UnsafePointer<CChar>?, UnsafeMutableRawPointer?) -> Void

@_cdecl("AppleAISessionRespondStream")
public func AppleAISessionRespondStream(
    sessionPtr: UnsafeMutableRawPointer, 
    prompt: UnsafePointer<CChar>, 
    callback: AppleAITokenCallback, 
    userData: UnsafeMutableRawPointer?
) -> UnsafeMutablePointer<CChar>? {
    let session = Unmanaged<AppleAISession>.fromOpaque(sessionPtr).takeUnretainedValue()
    let promptStr = String(cString: prompt)
    
    let semaphore = DispatchSemaphore(value: 0)
    var errorStr: String? = nil
    
    Task {
        do {
            let stream = session.session.streamResponse(to: promptStr)
            for try await partialResponse in stream {
                let token = partialResponse.content
                token.withCString { cStr in
                    callback(cStr, userData)
                }
            }
        } catch {
            errorStr = "ERROR: \(error.localizedDescription)"
        }
        semaphore.signal()
    }
    
    semaphore.wait()
    if let errorStr = errorStr {
        return strdup(errorStr)
    }
    return nil
}

@_cdecl("FreeAppleAIString")
public func FreeAppleAIString(str: UnsafeMutablePointer<CChar>?) {
    if let str = str {
        free(str)
    }
}
