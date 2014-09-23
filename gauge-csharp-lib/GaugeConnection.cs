using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Net;
using System.Net.Sockets;
using Google.ProtocolBuffers;
using main;

namespace gauge_csharp_lib
{
    public class GaugeConnection
    {
        private readonly TcpClient _tcpCilent;

        public GaugeConnection(int port)
        {
            var tcpClient = new TcpClient();
            try
            {
                tcpClient.Connect(new IPEndPoint(IPAddress.Loopback, port));
            }
            catch (Exception e)
            {
                throw new Exception("Could not connect", e);
            }
            _tcpCilent = tcpClient;
        }

        private static APIMessage ReadMessage(NetworkStream networkStream)
        {
            var responseBytes = ReadBytes(networkStream);
            return APIMessage.ParseFrom(responseBytes.ToArray());
        }

        private static void WriteApiMessage(IMessageLite apiRequest, Stream networkStream)
        {
            var bytes = apiRequest.ToByteArray();
            var cos = CodedOutputStream.CreateInstance(networkStream);
            cos.WriteRawVarint64((ulong) bytes.Length);
            cos.Flush();
            networkStream.Write(bytes, 0, bytes.Length);
            networkStream.Flush();
        }

        private static IEnumerable<byte> ReadBytes(Stream networkStream)
        {
            var codedInputStream = CodedInputStream.CreateInstance(networkStream);
            var messageLength = codedInputStream.ReadRawVarint64();
            for (ulong i = 0; i < messageLength; i++)
            {
                yield return codedInputStream.ReadRawByte();
            }
        }

        public IEnumerable<string> GetStepValue(IEnumerable<string> stepTexts, bool hasInlineTable)
        {
            foreach (var stepText in stepTexts)
            {
                var stepValueRequest = GetStepValueRequest.CreateBuilder()
                    .SetStepText(stepText)
                    .SetHasInlineTable(hasInlineTable)
                    .Build();
                var stepValueRequestMessage = APIMessage.CreateBuilder()
                    .SetMessageId(GenerateMessageId())
                    .SetMessageType(APIMessage.Types.APIMessageType.GetStepValueRequest)
                    .SetStepValueRequest(stepValueRequest)
                    .Build();
                var networkStream = _tcpCilent.GetStream();
                var apiMessage = WriteAndReadApiMessage(stepValueRequestMessage, networkStream);
                yield return apiMessage.StepValueResponse.StepValue.StepValue;
            }
        }

        private long GenerateMessageId()
        {
            return DateTime.Now.Ticks/TimeSpan.TicksPerMillisecond;
        }

        private APIMessage WriteAndReadApiMessage(APIMessage stepValueRequestMessage, NetworkStream networkStream)
        {
            lock (_tcpCilent)
            {
                WriteApiMessage(stepValueRequestMessage, networkStream);
                return ReadMessage(networkStream);
            }
        }
    }
}